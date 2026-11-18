// SPDX-License-Identifier: Apache-2.0
//! [Bounty #37] (`hanzoai/agents#87`): **Automatic prompt and workflow optimization** (DSPy).
//!
//! Clean-room simulation harness for replacing hand-written prompts with a
//! compile step that searches instructions and few-shot demonstrations against a
//! metric.
//!
//! # Design
//!
//! A task has a latent "alignment" landscape over a tuple of knobs (instruction
//! style, demonstration set, decoding hint). A good region exists, quality
//! decays smoothly with distance to it, and measurements are noisy.
//!
//! - `ManualPrompt`: one hand-written prompt, no search.
//! - `RandomSearch`: sample `N` candidates uniformly, keep the best on dev.
//! - `BootstrapCompile`: bootstrap candidates from the space, then walk each
//!   knob down the slope with shrinking steps, and re-measure the finalists
//!   with equal sample counts so noise cannot pick the winner.
//!
//! Metrics: dev accuracy, held-out test accuracy (overfit gap), candidates
//! evaluated, and the number of trials to reach a target.
//!
//! # Honest boundaries
//! - No LLM is called; accuracy is a modelled function of position in a synthetic
//!   landscape.
//! - The landscape's smoothness and noise are hand-set, not measured from a real
//!   DSPy compile.
//! - Candidate evaluation cost is a constant, not measured tokens.
//! - All figures are simulation outputs on the authoring host, NOT Hanzo
//!   expectations until reproduced with real models and a real metric.

use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum DspyError {
    #[error("invalid configuration: {0}")]
    InvalidConfig(String),
}

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

/// Values per knob. A candidate is a tuple of knob settings, not a scalar.
const SIDE: usize = 64;
/// Number of knobs: instruction style, demonstration set, decoding hint.
const DIMS: usize = 3;
/// Number of distinct configurations: `SIDE^DIMS`.
pub const SPACE: usize = SIDE * SIDE * SIDE;

/// The optimisation landscape: quality of a candidate configuration.
///
/// A candidate is a `DIMS`-tuple of knob settings, and quality decays with the
/// squared distance to one good region. This matters: uniform search over a
/// multi-dimensional space covers a vanishing fraction of it, while a
/// coordinate-wise walk can still follow the slope. A one-dimensional
/// landscape would make the comparison meaningless, because 60 uniform samples
/// already land next to the optimum and random search would look as good as
/// compilation.
pub struct Landscape {
    /// Quality of the hand-written prompt (the all-zero configuration).
    pub manual_quality: f64,
    /// Number of distinct candidate configurations.
    pub space: usize,
    /// How much a perfect configuration gains over the hand-written prompt.
    pub spread: f64,
    /// Measurement noise on a single dev evaluation.
    pub noise: f64,
    /// Knob settings of the good region, each in `0..1`.
    pub target: [f64; DIMS],
    /// Width of the good region, in knob units.
    pub width: f64,
}

impl Default for Landscape {
    fn default() -> Self {
        Self {
            manual_quality: 0.61,
            space: SPACE,
            spread: 0.28,
            noise: 0.02,
            target: [0.72, 0.35, 0.60],
            width: 0.18,
        }
    }
}

impl Landscape {
    /// Knob settings of candidate `i`.
    pub fn digits(i: usize) -> [usize; DIMS] {
        let mut rest = i.min(SPACE - 1);
        let mut d = [0usize; DIMS];
        for slot in d.iter_mut() {
            *slot = rest % SIDE;
            rest /= SIDE;
        }
        d
    }

    /// Candidate index for a tuple of knob settings.
    pub fn index(d: [usize; DIMS]) -> usize {
        d[0] + d[1] * SIDE + d[2] * SIDE * SIDE
    }

    /// The configuration closest to the good region.
    pub fn optimum(&self) -> usize {
        let mut d = [0usize; DIMS];
        for (slot, t) in d.iter_mut().zip(self.target.iter()) {
            *slot = (t * (SIDE - 1) as f64).round() as usize;
        }
        Self::index(d)
    }

    /// True quality of candidate `i`.
    pub fn quality(&self, i: usize) -> f64 {
        let mut dist2 = 0.0;
        for (v, t) in Self::digits(i).iter().zip(self.target.iter()) {
            let x = *v as f64 / (SIDE - 1) as f64;
            dist2 += ((x - t) / self.width).powi(2);
        }
        (self.manual_quality + self.spread * (-dist2).exp()).clamp(0.0, 0.99)
    }
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Policy {
    ManualPrompt,
    RandomSearch,
    BootstrapCompile,
}

impl Policy {
    pub fn label(&self) -> &'static str {
        match self {
            Policy::ManualPrompt => "manual-prompt",
            Policy::RandomSearch => "random-search",
            Policy::BootstrapCompile => "bootstrap-compile",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyResult {
    pub policy: String,
    pub candidates_evaluated: usize,
    pub dev_accuracy: f64,
    pub test_accuracy: f64,
    /// dev - test. Positive means overfitting to the dev set.
    pub overfit_gap: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BenchReport {
    pub bounty: String,
    pub hip: String,
    pub honesty_labels: Vec<String>,
    pub config: BenchConfig,
    pub results: Vec<PolicyResult>,
    pub notes: Vec<String>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BenchConfig {
    pub seed: u64,
    /// Search budget, in candidate evaluations.
    pub budget: usize,
    /// Size of the dev set used for selection.
    pub dev_size: usize,
    /// Size of the held-out test set.
    pub test_size: usize,
}

impl Default for BenchConfig {
    fn default() -> Self {
        Self { seed: 0xD5_9E11, budget: 120, dev_size: 200, test_size: 2_000 }
    }
}
impl BenchConfig {
    pub fn test_config() -> Self {
        Self { seed: 7, budget: 60, dev_size: 150, test_size: 1_000 }
    }
}

/// Average of `n` noisy measurements of a true quality.
fn measured(rng: &mut SplitMix64, truth: f64, noise: f64, n: usize) -> f64 {
    if n == 0 {
        return truth;
    }
    let mut acc = 0.0;
    for _ in 0..n {
        acc += truth + (rng.next_f64() - 0.5) * 2.0 * noise;
    }
    (acc / n as f64).clamp(0.0, 1.0)
}

pub fn evaluate(policy: Policy, land: &Landscape, cfg: &BenchConfig) -> PolicyResult {
    let mut rng = SplitMix64::new(cfg.seed);
    // Standard error of a dev estimate of the given size.
    let dev_noise = land.noise * (200.0 / cfg.dev_size.max(1) as f64).sqrt();

    let (best, evaluated) = match policy {
        Policy::ManualPrompt => (0usize, 1usize),
        Policy::RandomSearch => {
            let mut best = 0usize;
            let mut best_dev = f64::NEG_INFINITY;
            for _ in 0..cfg.budget {
                let i = (rng.next_u64() as usize) % land.space;
                let dev = measured(&mut rng, land.quality(i), dev_noise, 1);
                if dev > best_dev {
                    best_dev = dev;
                    best = i;
                }
            }
            (best, cfg.budget)
        }
        Policy::BootstrapCompile => {
            // Phase 1: bootstrap candidates, then walk each knob down the slope
            // (coordinate descent) with shrinking steps. Phase 2: re-measure the
            // finalists with equal sample counts so a single lucky estimate
            // cannot win on noise alone.
            let refine_slots = (cfg.budget / 4).clamp(2, 18);
            let search_budget = cfg.budget.saturating_sub(refine_slots).max(6);
            let mut seen: Vec<(usize, f64)> = Vec::new();
            let mut best = 0usize;
            let mut best_dev = f64::NEG_INFINITY;

            let warmup = (search_budget / 2).max(3).min(search_budget);
            for _ in 0..warmup {
                let i = (rng.next_u64() as usize) % land.space;
                let dev = measured(&mut rng, land.quality(i), dev_noise, 2);
                seen.push((i, dev));
                if dev > best_dev {
                    best_dev = dev;
                    best = i;
                }
            }

            let mut spent = warmup;
            for step in [SIDE / 4, SIDE / 8, SIDE / 16, 1] {
                // Only start a pass that fits: half-finished passes compare
                // knobs on unequal evidence.
                if spent + 2 * DIMS > search_budget {
                    break;
                }
                for dim in 0..DIMS {
                    for dir in [-1i32, 1i32] {
                        let mut d = Landscape::digits(best);
                        let v = d[dim] as i32 + dir * step as i32;
                        if v < 0 || v >= SIDE as i32 {
                            continue;
                        }
                        d[dim] = v as usize;
                        let cand = Landscape::index(d);
                        if cand == best {
                            continue;
                        }
                        let dev = measured(&mut rng, land.quality(cand), dev_noise, 2);
                        seen.push((cand, dev));
                        spent += 1;
                        if dev > best_dev {
                            best_dev = dev;
                            best = cand;
                        }
                    }
                }
            }

            // Phase 2: re-measure the top few distinct candidates with the same
            // number of samples each, then keep the winner.
            seen.sort_by(|a, b| b.1.partial_cmp(&a.1).unwrap_or(std::cmp::Ordering::Equal));
            let mut top: Vec<usize> = Vec::new();
            for (idx, _) in seen.iter() {
                if !top.contains(idx) {
                    top.push(*idx);
                }
                if top.len() == 3 {
                    break;
                }
            }
            if !top.contains(&best) {
                top.push(best);
            }
            let per = (refine_slots / top.len()).max(1);
            let mut refined_best = best;
            let mut refined_dev = f64::NEG_INFINITY;
            for idx in top {
                if spent + per > cfg.budget {
                    break;
                }
                let dev = measured(&mut rng, land.quality(idx), dev_noise, per);
                spent += per;
                if dev > refined_dev {
                    refined_dev = dev;
                    refined_best = idx;
                }
            }
            (refined_best, spent)
        }
    };

    let truth = land.quality(best);
    let dev_accuracy = measured(&mut rng, truth, dev_noise, 3);
    // Test estimate on a much larger held-out set: almost no noise.
    let test_accuracy = measured(&mut rng, truth, land.noise * (200.0 / cfg.test_size.max(1) as f64).sqrt(), 3);

    PolicyResult {
        policy: policy.label().into(),
        candidates_evaluated: evaluated,
        dev_accuracy,
        test_accuracy,
        overfit_gap: dev_accuracy - test_accuracy,
    }
}

pub fn run_bench(cfg: &BenchConfig) -> Result<BenchReport, DspyError> {
    if cfg.budget == 0 {
        return Err(DspyError::InvalidConfig("budget must be positive".into()));
    }
    if cfg.dev_size == 0 || cfg.test_size == 0 {
        return Err(DspyError::InvalidConfig("dev_size and test_size must be positive".into()));
    }
    let land = Landscape::default();
    let results = vec![
        evaluate(Policy::ManualPrompt, &land, cfg),
        evaluate(Policy::RandomSearch, &land, cfg),
        evaluate(Policy::BootstrapCompile, &land, cfg),
    ];

    let mut notes = Vec::new();
    if let (Some(manual), Some(boot)) = (
        results.iter().find(|r| r.policy == "manual-prompt"),
        results.iter().find(|r| r.policy == "bootstrap-compile"),
    ) {
        notes.push(format!(
            "compiling lifts held-out accuracy {:.1}% -> {:.1}% using {} candidate evaluations; random search of the same budget reaches {:.1}%",
            manual.test_accuracy * 100.0,
            boot.test_accuracy * 100.0,
            boot.candidates_evaluated,
            results
                .iter()
                .find(|r| r.policy == "random-search")
                .map(|r| r.test_accuracy * 100.0)
                .unwrap_or(0.0)
        ));
    }

    Ok(BenchReport {
        bounty: "hanzoai/agents#87 [Bounty #37] Automatic prompt and workflow optimization (DSPy)".into(),
        hip: "HIP-1348: Reliable, Economical and Verifiable AI Infrastructure".into(),
        honesty_labels: vec![
            "no LLM is called - accuracy is a modelled function of position in a synthetic landscape".into(),
            "landscape smoothness and measurement noise are hand-set".into(),
            "candidate evaluation cost is a constant, not measured tokens".into(),
            "All figures are simulation outputs on the authoring host, NOT Hanzo expectations until reproduced with real models and a real metric.".into(),
        ],
        config: cfg.clone(),
        results,
        notes,
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

    fn find<'a>(r: &'a [PolicyResult], p: &str) -> &'a PolicyResult {
        r.iter().find(|x| x.policy == p).expect(p)
    }

    #[test]
    fn compiling_beats_manual() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let manual = find(&report.results, "manual-prompt");
        let boot = find(&report.results, "bootstrap-compile");
        assert!(
            boot.test_accuracy > manual.test_accuracy + 0.10,
            "compile {:.3} should beat manual {:.3} by a clear margin",
            boot.test_accuracy,
            manual.test_accuracy
        );
    }

    #[test]
    fn bootstrap_beats_random_search_on_equal_budget() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let rs = find(&report.results, "random-search");
        let boot = find(&report.results, "bootstrap-compile");
        assert!(
            boot.test_accuracy > rs.test_accuracy + 0.05,
            "bootstrap {:.3} should beat random {:.3} on the same budget",
            boot.test_accuracy,
            rs.test_accuracy
        );
        assert!(boot.candidates_evaluated <= 60);
        assert!(rs.candidates_evaluated <= 60);
    }

    #[test]
    fn manual_prompt_costs_one_evaluation() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        assert_eq!(find(&report.results, "manual-prompt").candidates_evaluated, 1);
    }

    #[test]
    fn landscape_has_single_peak() {
        let land = Landscape::default();
        let peak = land.optimum();
        let best = land.quality(peak);
        assert!(
            (best - (land.manual_quality + land.spread)).abs() < 1e-3,
            "peak quality {:.4} should be the top of the landscape",
            best
        );
        // Every one-knob neighbour is worse: the peak is a true local maximum.
        let d = Landscape::digits(peak);
        for dim in 0..DIMS {
            for delta in [-1i32, 1] {
                let v = d[dim] as i32 + delta;
                if v < 0 || v >= SIDE as i32 {
                    continue;
                }
                let mut n = d;
                n[dim] = v as usize;
                assert!(land.quality(Landscape::index(n)) < best);
            }
        }
        // The hand-written prompt sits at the floor of the landscape.
        assert!(land.quality(0) <= land.manual_quality + 1e-9);
    }

    #[test]
    fn accuracies_are_bounded() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        for r in &report.results {
            assert!(r.dev_accuracy >= 0.0 && r.dev_accuracy <= 1.0);
            assert!(r.test_accuracy >= 0.0 && r.test_accuracy <= 1.0);
        }
    }

    #[test]
    fn bad_config_rejected() {
        let mut c = BenchConfig::test_config();
        c.budget = 0;
        assert!(run_bench(&c).is_err());
        let mut c = BenchConfig::test_config();
        c.dev_size = 0;
        assert!(run_bench(&c).is_err());
    }
}
