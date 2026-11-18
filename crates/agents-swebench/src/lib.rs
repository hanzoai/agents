// SPDX-License-Identifier: Apache-2.0
//! [Bounty #31] (`hanzoai/agents#84`): **Long-horizon coding-agent evaluation** (SWE-bench style).
//!
//! Clean-room evaluation harness measuring how agent resolve-rate degrades as the
//! horizon grows: context window pressure, iteration budget, and repo scale.
//!
//! # Design
//!
//! - Deterministic synthetic repo instances with a known ground-truth patch site.
//! - Agent policies differ in **context management**: full-history (baseline) vs
//!   rolling-summary vs retrieval-of-relevant-files.
//! - Resolve is scored pass/fail per instance; we report resolve-rate and
//!   cost-per-resolve across horizon buckets.
//!
//! # Honest boundaries
//! - No real LLM is invoked. Resolution probability is a deterministic function of
//!   context utilisation and iteration budget, calibrated by hand to match the
//!   qualitative shape reported in the SWE-bench literature (long-context degradation).
//! - Synthetic repos, synthetic patches — not real GitHub issue/PR pairs.
//! - All figures are simulation outputs on the authoring host, NOT Hanzo expectations
//!   until reproduced against the real SWE-bench Verified harness with a live model.

use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use thiserror::Error;

#[derive(Debug, Error)]
pub enum SWEError {
    #[error("invalid configuration: {0}")]
    InvalidConfig(String),
}

/// Deterministic PRNG (SplitMix64).
#[derive(Debug, Clone)]
pub struct SplitMix64 {
    state: u64,
}
impl SplitMix64 {
    pub fn new(seed: u64) -> Self {
        Self { state: seed }
    }
    pub fn next_u64(&mut self) -> u64 {
        self.state = self.state.wrapping_add(0x9E37_79B9_7F4A_7C15);
        let mut z = self.state;
        z = (z ^ (z >> 30)).wrapping_mul(0xBF58_476D_1CE4_E5B9);
        z = (z ^ (z >> 27)).wrapping_mul(0x94D0_49BB_1331_11EB);
        z ^ (z >> 31)
    }
    pub fn next_f64(&mut self) -> f64 {
        (self.next_u64() >> 11) as f64 / (1u64 << 53) as f64
    }
}

/// A synthetic coding task.
#[derive(Debug, Clone)]
pub struct Task {
    /// Number of source files the agent may need to read.
    pub repo_files: usize,
    /// Lines of code in the repo (proxy for exploration cost).
    pub repo_loc: usize,
    /// How many files must actually be inspected to find the fix site.
    pub files_to_inspect: usize,
    /// Minimum reasoning iterations required for a correct patch.
    pub min_iterations: usize,
    /// Where in the repo the fix site lives (0..repo_files).
    pub fix_site: usize,
}

/// Horizon buckets we report separately.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
pub enum Horizon {
    Short,  // <= 10 files, <= 2k LOC
    Medium, // <= 100 files, <= 20k LOC
    Long,   // > 100 files, > 20k LOC
}
impl Horizon {
    pub fn of(t: &Task) -> Self {
        if t.repo_files <= 10 && t.repo_loc <= 2_000 {
            Horizon::Short
        } else if t.repo_files <= 100 && t.repo_loc <= 20_000 {
            Horizon::Medium
        } else {
            Horizon::Long
        }
    }
    pub fn label(&self) -> &'static str {
        match self {
            Horizon::Short => "short",
            Horizon::Medium => "medium",
            Horizon::Long => "long",
        }
    }
}

pub fn generate_tasks(seed: u64, n: usize) -> Vec<Task> {
    let mut rng = SplitMix64::new(seed);
    (0..n)
        .map(|_| {
            // Log-uniform repo scale across short/medium/long horizons.
            let scale = rng.next_f64();
            let (files, loc) = if scale < 0.4 {
                (3 + (rng.next_u64() as usize % 8), 500 + (rng.next_u64() as usize % 1_500))
            } else if scale < 0.8 {
                (20 + (rng.next_u64() as usize % 80), 5_000 + (rng.next_u64() as usize % 15_000))
            } else {
                (150 + (rng.next_u64() as usize % 850), 25_000 + (rng.next_u64() as usize % 75_000))
            };
            let files_to_inspect = 1 + (rng.next_u64() as usize % files.max(1).min(6));
            let min_iterations = 1 + (rng.next_u64() as usize % 8);
            let fix_site = rng.next_u64() as usize % files.max(1);
            Task { repo_files: files, repo_loc: loc, files_to_inspect, min_iterations, fix_site }
        })
        .collect()
}

/// How the agent manages its context as the horizon grows.
#[derive(Debug, Clone, Copy, PartialEq)]
pub enum ContextPolicy {
    /// Keep everything. Context fills up; recall collapses.
    FullHistory,
    /// Keep a rolling summary of the last N observations.
    RollingSummary { window: usize },
    /// Retrieve only files whose path/contents look relevant to the issue.
    Retrieval { recall: f64 },
}

impl ContextPolicy {
    pub fn label(&self) -> String {
        match self {
            ContextPolicy::FullHistory => "full-history".into(),
            ContextPolicy::RollingSummary { window } => format!("rolling-summary(window={})", window),
            ContextPolicy::Retrieval { recall } => format!("retrieval(recall={:.2})", recall),
        }
    }

    /// Fraction of the needed fix-site evidence that survives into the agent's
    /// working context. This is the modelled source of horizon degradation.
    fn evidence_retention(&self, t: &Task) -> f64 {
        let horizon_pressure = (t.repo_loc as f64 / 25_000.0).clamp(0.0, 1.0);
        match self {
            // Everything kept, but attention dilutes with context length.
            ContextPolicy::FullHistory => 1.0 - 0.65 * horizon_pressure,
            // Summary keeps recent window; older fix sites get summarised away.
            ContextPolicy::RollingSummary { window } => {
                let covered = (*window as f64 / t.files_to_inspect.max(1) as f64).min(1.0);
                (0.45 + 0.5 * covered) * (1.0 - 0.2 * horizon_pressure)
            }
            // Retrieval recall is horizon-independent by construction.
            ContextPolicy::Retrieval { recall } => *recall,
        }
    }

    /// Exploration cost in "iterations" — how long it takes to locate the fix site.
    fn exploration_cost(&self, t: &Task) -> f64 {
        match self {
            ContextPolicy::FullHistory => t.repo_files as f64 * 0.15 + t.files_to_inspect as f64,
            ContextPolicy::RollingSummary { .. } => t.repo_files as f64 * 0.08 + t.files_to_inspect as f64,
            ContextPolicy::Retrieval { recall } => {
                // Retrieval finds the right file quickly, but if it misses we must
                // fall back to linear scanning.
                let p_hit = recall.clamp(0.0, 1.0);
                p_hit * 1.5 + (1.0 - p_hit) * (t.repo_files as f64 * 0.15)
            }
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct HorizonStats {
    pub horizon: String,
    pub tasks: usize,
    pub resolved: usize,
    pub resolve_rate: f64,
    pub mean_iterations_used: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyResult {
    pub policy: String,
    pub iteration_budget: usize,
    pub resolve_rate: f64,
    pub cost_per_resolve: f64,
    pub by_horizon: Vec<HorizonStats>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BenchReport {
    pub bounty: String,
    pub hip: String,
    pub honesty_labels: Vec<String>,
    pub config: BenchConfig,
    pub results: Vec<PolicyResult>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BenchConfig {
    pub seed: u64,
    pub n_tasks: usize,
    pub iteration_budget: usize,
}

impl Default for BenchConfig {
    fn default() -> Self {
        Self { seed: 0x5EE_B0C, n_tasks: 400, iteration_budget: 12 }
    }
}
impl BenchConfig {
    pub fn test_config() -> Self {
        Self { seed: 7, n_tasks: 60, iteration_budget: 12 }
    }
}

/// Evaluate one policy over the task set.
pub fn evaluate(policy: ContextPolicy, tasks: &[Task], budget: usize) -> PolicyResult {
    let mut by_horizon: HashMap<Horizon, (usize, usize, f64)> = HashMap::new();
    let mut resolved_total = 0usize;
    let mut iteration_total = 0.0f64;

    for t in tasks {
        let retention = policy.evidence_retention(t).clamp(0.0, 1.0);
        let cost = policy.exploration_cost(t);
        // Iterations actually consumed = exploration + reasoning, capped by budget.
        let needed = cost + t.min_iterations as f64;
        let used = needed.min(budget as f64);
        iteration_total += used;

        // Success requires (a) enough budget to finish and (b) the evidence survived.
        // Retention scales the success probability smoothly rather than as a hard cut.
        let budget_ok = needed <= budget as f64;
        let success_prob = if budget_ok { retention } else { retention * 0.25 };
        // Deterministic pseudo-verdict from the task shape so runs are reproducible.
        let draw = ((t.repo_loc * 31 + t.fix_site * 17 + t.files_to_inspect * 7) % 1000) as f64 / 1000.0;
        let ok = draw < success_prob;

        let e = by_horizon.entry(Horizon::of(t)).or_insert((0, 0, 0.0));
        e.0 += 1;
        if ok {
            e.1 += 1;
            resolved_total += 1;
        }
        e.2 += used;
    }

    let mut horizons: Vec<HorizonStats> = by_horizon
        .into_iter()
        .map(|(h, (n, ok, iters))| HorizonStats {
            horizon: h.label().to_string(),
            tasks: n,
            resolved: ok,
            resolve_rate: ok as f64 / n.max(1) as f64,
            mean_iterations_used: iters / n.max(1) as f64,
        })
        .collect();
    horizons.sort_by(|a, b| a.horizon.cmp(&b.horizon));

    let n = tasks.len().max(1) as f64;
    PolicyResult {
        policy: policy.label(),
        iteration_budget: budget,
        resolve_rate: resolved_total as f64 / n,
        cost_per_resolve: if resolved_total > 0 { iteration_total / resolved_total as f64 } else { f64::INFINITY },
        by_horizon: horizons,
    }
}

pub fn run_bench(cfg: &BenchConfig) -> Result<BenchReport, SWEError> {
    if cfg.n_tasks == 0 {
        return Err(SWEError::InvalidConfig("n_tasks must be positive".into()));
    }
    if cfg.iteration_budget == 0 {
        return Err(SWEError::InvalidConfig("iteration_budget must be positive".into()));
    }
    let tasks = generate_tasks(cfg.seed, cfg.n_tasks);
    let policies = [
        ContextPolicy::FullHistory,
        ContextPolicy::RollingSummary { window: 3 },
        ContextPolicy::Retrieval { recall: 0.85 },
    ];
    let results = policies
        .iter()
        .map(|p| evaluate(*p, &tasks, cfg.iteration_budget))
        .collect();

    Ok(BenchReport {
        bounty: "hanzoai/agents#84 [Bounty #31] Long-horizon coding-agent evaluation (SWE-bench style)".into(),
        hip: "HIP-1346: Evidence-Grounded Retrieval and Decision Evaluation".into(),
        honesty_labels: vec![
            "no real LLM invoked — resolution probability is a deterministic model of context pressure".into(),
            "synthetic repositories and patches, not real GitHub issue/PR pairs".into(),
            "retention/exploration curves are hand-calibrated to the qualitative shape of SWE-bench findings".into(),
            "All figures are simulation outputs on the authoring host, NOT Hanzo expectations until reproduced on the real SWE-bench harness with a live model.".into(),
        ],
        config: cfg.clone(),
        results,
    })
}

pub fn write_results_json(report: &BenchReport, path: &std::path::Path) -> Result<(), Box<dyn std::error::Error>> {
    if let Some(parent) = path.parent() {
        std::fs::create_dir_all(parent)?;
    }
    std::fs::write(path, serde_json::to_string_pretty(report)?)?;
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn all_horizons_are_populated() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        for r in &report.results {
            assert!(!r.by_horizon.is_empty(), "{} produced no horizon buckets", r.policy);
        }
    }

    #[test]
    fn retrieval_beats_full_history_on_long_horizon() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let full = report.results.iter().find(|r| r.policy.starts_with("full-history")).unwrap();
        let ret = report.results.iter().find(|r| r.policy.starts_with("retrieval")).unwrap();
        let long = |r: &PolicyResult| {
            r.by_horizon.iter().find(|h| h.horizon == "long").map(|h| h.resolve_rate).unwrap_or(0.0)
        };
        assert!(
            long(ret) >= long(full),
            "retrieval {:.3} should not lose to full-history {:.3} on long horizon",
            long(ret),
            long(full)
        );
    }

    #[test]
    fn full_history_degrades_with_horizon() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let full = report.results.iter().find(|r| r.policy.starts_with("full-history")).unwrap();
        let get = |name: &str| {
            full.by_horizon.iter().find(|h| h.horizon == name).map(|h| h.resolve_rate).unwrap_or(0.0)
        };
        assert!(
            get("short") >= get("long"),
            "full-history short {:.3} should be >= long {:.3}",
            get("short"),
            get("long")
        );
    }

    #[test]
    fn zero_tasks_rejected() {
        assert!(run_bench(&BenchConfig { seed: 1, n_tasks: 0, iteration_budget: 5 }).is_err());
    }

    #[test]
    fn results_are_reproducible() {
        let cfg = BenchConfig::test_config();
        let a = run_bench(&cfg).unwrap();
        let b = run_bench(&cfg).unwrap();
        assert_eq!(a.results.len(), b.results.len());
        for (x, y) in a.results.iter().zip(b.results.iter()) {
            assert_eq!(x.resolve_rate, y.resolve_rate);
        }
    }
}
