// SPDX-License-Identifier: Apache-2.0
//! [Bounty #40] (`hanzoai/agents#89`): **Computer-workflow benchmark on Visor** (OSWorld).
//!
//! Clean-room simulation harness for grounding desktop actions — the step where
//! a computer-use agent turns "click the Save button" into a concrete action.
//!
//! # Design
//!
//! A desktop is a tree of UI elements, some of which are labelled in the
//! accessibility tree, some only visible as pixels. Each task is a sequence of
//! targets. Grounding strategies:
//!
//! - `PixelClick`: predict coordinates from the screenshot only.
//! - `A11yTree`: resolve targets through the accessibility tree.
//! - `HybridGrounded`: accessibility tree first, pixel fallback for unlabelled
//!   targets, plus a post-action verification step that detects and retries a
//!   failed action.
//!
//! Metrics: task success, mean steps per task, misclick rate, and how many
//! misclicks are recovered before the task fails.
//!
//! # Honest boundaries
//! - No OS, screenshot, or accessibility API is exercised; targets are synthetic.
//! - Hit rates per strategy are hand-set, not measured from a real OSWorld run.
//! - Verification is assumed to cost one extra step and to be imperfect.
//! - All figures are simulation outputs on the authoring host, NOT Hanzo
//!   expectations until reproduced on a real Visor/OSWorld environment.

use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum OsworldError {
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

/// A target element on screen.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Target {
    /// Labelled in the accessibility tree and visually distinct.
    Labelled,
    /// Visible but unlabelled (canvas, custom widget, icon-only button).
    Unlabelled,
    /// Off-screen until scrolled; requires a navigation step first.
    Offscreen,
}

impl Target {
    pub fn name(&self) -> &'static str {
        match self {
            Target::Labelled => "labelled",
            Target::Unlabelled => "unlabelled",
            Target::Offscreen => "offscreen",
        }
    }
}

#[derive(Debug, Clone)]
pub struct Task {
    pub id: u32,
    pub targets: Vec<Target>,
}

pub fn generate_tasks(seed: u64, n_tasks: usize, targets_per_task: usize, unlabelled_share: f64, offscreen_share: f64) -> Vec<Task> {
    let mut rng = SplitMix64::new(seed);
    (0..n_tasks)
        .map(|i| {
            let targets = (0..targets_per_task)
                .map(|_| {
                    let r = rng.next_f64();
                    if r < offscreen_share {
                        Target::Offscreen
                    } else if r < offscreen_share + unlabelled_share {
                        Target::Unlabelled
                    } else {
                        Target::Labelled
                    }
                })
                .collect();
            Task { id: i as u32, targets }
        })
        .collect()
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Policy {
    PixelClick,
    A11yTree,
    HybridGrounded,
}

impl Policy {
    pub fn label(&self) -> &'static str {
        match self {
            Policy::PixelClick => "pixel-click",
            Policy::A11yTree => "a11y-tree",
            Policy::HybridGrounded => "hybrid-grounded",
        }
    }

    /// Probability the first attempt lands on the right element.
    pub fn first_attempt_hit(&self, t: Target) -> f64 {
        match (self, t) {
            (Policy::PixelClick, Target::Labelled) => 0.72,
            (Policy::PixelClick, Target::Unlabelled) => 0.58,
            (Policy::PixelClick, Target::Offscreen) => 0.20,
            (Policy::A11yTree, Target::Labelled) => 0.97,
            (Policy::A11yTree, Target::Unlabelled) => 0.24,
            (Policy::A11yTree, Target::Offscreen) => 0.55,
            (Policy::HybridGrounded, Target::Labelled) => 0.97,
            (Policy::HybridGrounded, Target::Unlabelled) => 0.71,
            (Policy::HybridGrounded, Target::Offscreen) => 0.62,
        }
    }

    /// Whether the policy verifies the result and retries a failed action.
    pub fn verifies(&self) -> bool {
        matches!(self, Policy::HybridGrounded)
    }

    /// Probability that verification actually notices a failed action.
    pub fn detection_rate(&self) -> f64 {
        if self.verifies() {
            0.85
        } else {
            0.0
        }
    }

    /// Extra steps spent per target on verification.
    pub fn verification_steps(&self) -> usize {
        if self.verifies() {
            1
        } else {
            0
        }
    }

    /// Probability a retry succeeds after a detected failure.
    pub fn retry_success(&self) -> f64 {
        0.55
    }

    /// Max retries per target.
    pub fn max_retries(&self) -> usize {
        if self.verifies() {
            1
        } else {
            0
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyResult {
    pub policy: String,
    pub tasks: usize,
    pub task_success_rate: f64,
    pub misclicks: usize,
    pub misclick_rate: f64,
    pub recovered_misclicks: usize,
    pub steps_per_task: f64,
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
    pub n_tasks: usize,
    pub targets_per_task: usize,
    pub unlabelled_share: f64,
    pub offscreen_share: f64,
}

impl Default for BenchConfig {
    fn default() -> Self {
        Self {
            seed: 0x05_7E12,
            n_tasks: 1_500,
            targets_per_task: 9,
            unlabelled_share: 0.28,
            offscreen_share: 0.12,
        }
    }
}
impl BenchConfig {
    pub fn test_config() -> Self {
        Self { seed: 7, n_tasks: 300, targets_per_task: 9, unlabelled_share: 0.28, offscreen_share: 0.12 }
    }
}

pub fn evaluate(policy: Policy, tasks: &[Task], cfg: &BenchConfig) -> PolicyResult {
    let mut rng = SplitMix64::new(cfg.seed ^ 0x0517);
    let mut successes = 0usize;
    let mut misclicks = 0usize;
    let mut recovered = 0usize;
    let mut steps = 0usize;
    let mut attempts = 0usize;

    for t in tasks {
        let mut ok = true;
        for target in &t.targets {
            if !ok {
                break;
            }
            let mut done = false;
            let mut tries = 0usize;
            loop {
                tries += 1;
                attempts += 1;
                steps += 1;
                // Off-screen targets need a navigation step before the click.
                if *target == Target::Offscreen {
                    steps += 1;
                }
                let hit = rng.next_f64() < policy.first_attempt_hit(*target);
                if hit {
                    done = true;
                } else {
                    misclicks += 1;
                    if policy.verifies() && tries <= policy.max_retries() {
                        steps += policy.verification_steps();
                        // Verification must first notice the failure, then the
                        // retry must land.
                        if rng.next_f64() < policy.detection_rate()
                            && rng.next_f64() < policy.retry_success()
                        {
                            recovered += 1;
                            done = true;
                        }
                    }
                }
                if done || tries > policy.max_retries() + 1 {
                    break;
                }
            }
            if !done {
                ok = false;
            }
        }
        if ok {
            successes += 1;
        }
    }

    let n = tasks.len().max(1) as f64;
    PolicyResult {
        policy: policy.label().into(),
        tasks: tasks.len(),
        task_success_rate: successes as f64 / n,
        misclicks,
        misclick_rate: misclicks as f64 / attempts.max(1) as f64,
        recovered_misclicks: recovered,
        steps_per_task: steps as f64 / n,
    }
}

pub fn run_bench(cfg: &BenchConfig) -> Result<BenchReport, OsworldError> {
    if cfg.n_tasks == 0 {
        return Err(OsworldError::InvalidConfig("n_tasks must be positive".into()));
    }
    if cfg.targets_per_task == 0 {
        return Err(OsworldError::InvalidConfig("targets_per_task must be positive".into()));
    }
    if !(0.0..=1.0).contains(&cfg.unlabelled_share) || !(0.0..=1.0).contains(&cfg.offscreen_share) {
        return Err(OsworldError::InvalidConfig("shares must be in [0,1]".into()));
    }
    if cfg.unlabelled_share + cfg.offscreen_share > 1.0 {
        return Err(OsworldError::InvalidConfig("unlabelled_share + offscreen_share must be <= 1".into()));
    }
    let tasks = generate_tasks(cfg.seed, cfg.n_tasks, cfg.targets_per_task, cfg.unlabelled_share, cfg.offscreen_share);
    let results = vec![
        evaluate(Policy::PixelClick, &tasks, cfg),
        evaluate(Policy::A11yTree, &tasks, cfg),
        evaluate(Policy::HybridGrounded, &tasks, cfg),
    ];

    let mut notes = Vec::new();
    if let (Some(a11y), Some(hybrid)) = (
        results.iter().find(|r| r.policy == "a11y-tree"),
        results.iter().find(|r| r.policy == "hybrid-grounded"),
    ) {
        notes.push(format!(
            "hybrid grounding lifts task success from {:.1}% (a11y-only) and {:.1}% (pixel-only) to {:.1}%, recovering {} of {} misclicks, at {:.1} steps per task",
            a11y.task_success_rate * 100.0,
            results
                .iter()
                .find(|r| r.policy == "pixel-click")
                .map(|r| r.task_success_rate * 100.0)
                .unwrap_or(0.0),
            hybrid.task_success_rate * 100.0,
            hybrid.recovered_misclicks,
            hybrid.misclicks,
            hybrid.steps_per_task
        ));
    }

    Ok(BenchReport {
        bounty: "hanzoai/agents#89 [Bounty #40] Computer-workflow benchmark on Visor (OSWorld)".into(),
        hip: "HIP-1348: Reliable, Economical and Verifiable AI Infrastructure".into(),
        honesty_labels: vec![
            "no OS, screenshot or accessibility API is exercised - targets are synthetic".into(),
            "per-strategy hit rates are hand-set, not measured from a real OSWorld run".into(),
            "verification is assumed to cost one step, to notice 85% of failures, and to recover 55% of detected failures".into(),
            "All figures are simulation outputs on the authoring host, NOT Hanzo expectations until reproduced on a real Visor/OSWorld environment.".into(),
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
    fn hybrid_beats_both_single_modalities() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let hybrid = find(&report.results, "hybrid-grounded");
        for p in ["pixel-click", "a11y-tree"] {
            let other = find(&report.results, p);
            assert!(
                hybrid.task_success_rate > other.task_success_rate,
                "hybrid {:.3} should beat {} {:.3}",
                hybrid.task_success_rate,
                p,
                other.task_success_rate
            );
        }
    }

    #[test]
    fn only_verifying_policy_recovers_misclicks() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        assert_eq!(find(&report.results, "pixel-click").recovered_misclicks, 0);
        assert_eq!(find(&report.results, "a11y-tree").recovered_misclicks, 0);
        assert!(find(&report.results, "hybrid-grounded").recovered_misclicks > 0);
    }

    #[test]
    fn a11y_beats_pixel_on_labelled_targets() {
        assert!(Policy::A11yTree.first_attempt_hit(Target::Labelled) > Policy::PixelClick.first_attempt_hit(Target::Labelled));
    }

    #[test]
    fn hybrid_costs_more_steps() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let a11y = find(&report.results, "a11y-tree");
        let hybrid = find(&report.results, "hybrid-grounded");
        assert!(hybrid.steps_per_task > a11y.steps_per_task);
    }

    #[test]
    fn rates_are_bounded() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        for r in &report.results {
            assert!(r.task_success_rate >= 0.0 && r.task_success_rate <= 1.0);
            assert!(r.misclick_rate >= 0.0 && r.misclick_rate <= 1.0);
        }
    }

    #[test]
    fn bad_config_rejected() {
        let mut c = BenchConfig::test_config();
        c.n_tasks = 0;
        assert!(run_bench(&c).is_err());
        let mut c = BenchConfig::test_config();
        c.unlabelled_share = 0.8;
        c.offscreen_share = 0.5;
        assert!(run_bench(&c).is_err());
    }
}
