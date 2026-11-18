// SPDX-License-Identifier: Apache-2.0
//! [Bounty #36] (`hanzoai/agents#86`): **Learning from agent execution traces** (Agent Lightning).
//!
//! Clean-room simulation harness for turning raw agent trajectories into
//! training signal without rewriting the agent.
//!
//! # Design
//!
//! An agent run is a sequence of steps. Only the final outcome carries a reward,
//! and it is sparse. Three ways to learn from the same traces:
//!
//! - `OutcomeOnly`: treat the whole trajectory as one sample with the terminal
//!   reward (REINFORCE-style). Unbiased but very high variance.
//! - `UniformCredit`: assign the terminal reward to every step equally. Low
//!   variance, systematically wrong on long trajectories.
//! - `SegmentAdvantage`: split the trajectory at tool boundaries, estimate each
//!   segment's contribution from how the outcome distribution shifts, and weight
//!   steps by segment advantage.
//!
//! Metrics: episodes needed to reach a target success rate, final success rate at
//! a fixed episode budget, and the variance of the gradient estimator.
//!
//! # Honest boundaries
//! - No model is trained; the learning curve is a modelled function of estimator
//!   variance and credit-assignment bias.
//! - Variance and bias constants are hand-set, not measured from a real RL run.
//! - All figures are simulation outputs on the authoring host, NOT Hanzo
//!   expectations until reproduced with a real training loop.

use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum LightningError {
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

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Policy {
    OutcomeOnly,
    UniformCredit,
    SegmentAdvantage,
}

impl Policy {
    pub fn label(&self) -> &'static str {
        match self {
            Policy::OutcomeOnly => "outcome-only",
            Policy::UniformCredit => "uniform-credit",
            Policy::SegmentAdvantage => "segment-advantage",
        }
    }

    /// Standard deviation of the per-episode gradient estimate.
    pub fn gradient_noise(&self, steps: usize) -> f64 {
        let s = steps.max(1) as f64;
        match self {
            Policy::OutcomeOnly => 0.42 * s.sqrt(),
            Policy::UniformCredit => 0.30,
            Policy::SegmentAdvantage => 0.16 * s.sqrt() / 2.0,
        }
    }

    /// Systematic bias in the credit assigned to each step, in [0,1].
    /// 0 = unbiased, higher = worse.
    pub fn credit_bias(&self, steps: usize) -> f64 {
        let s = steps.max(1) as f64;
        match self {
            Policy::OutcomeOnly => 0.0,
            // Every step gets 1/steps of the credit: on long runs this badly
            // under-credits the decisive step.
            Policy::UniformCredit => 1.0 - 1.0 / s,
            // Segment boundaries recover most of the structure.
            Policy::SegmentAdvantage => 0.08,
        }
    }
}

/// Simulated learning curve: success rate after `episode` training episodes.
pub fn success_rate(policy: Policy, steps: usize, episode: usize, ceiling: f64) -> f64 {
    let noise = policy.gradient_noise(steps);
    let bias = policy.credit_bias(steps);
    // Effective progress per episode falls with noise and with bias (biased
    // gradients move in a partly wrong direction).
    let rate = 1.0 / (1.0 + 40.0 * noise + 12.0 * bias);
    let e = episode as f64;
    // Saturating curve towards (ceiling - bias penalty).
    let asymptote = (ceiling - 0.55 * bias).max(0.0);
    asymptote * (1.0 - (-rate * e / 20.0).exp())
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct CurvePoint {
    pub episode: usize,
    pub success_rate: f64,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyResult {
    pub policy: String,
    pub gradient_noise: f64,
    pub credit_bias: f64,
    /// First episode reaching the target, or null if never.
    pub episodes_to_target: Option<usize>,
    pub final_success_rate: f64,
    pub curve: Vec<CurvePoint>,
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
    pub max_episodes: usize,
    pub steps_per_trajectory: usize,
    pub target_success_rate: f64,
    /// Best achievable success rate for this task family.
    pub ceiling: f64,
}

impl Default for BenchConfig {
    fn default() -> Self {
        Self {
            seed: 0xA6E_771,
            max_episodes: 4_000,
            steps_per_trajectory: 24,
            target_success_rate: 0.60,
            ceiling: 0.82,
        }
    }
}
impl BenchConfig {
    pub fn test_config() -> Self {
        Self {
            seed: 7,
            max_episodes: 600,
            steps_per_trajectory: 24,
            target_success_rate: 0.60,
            ceiling: 0.82,
        }
    }
}

pub fn evaluate(policy: Policy, cfg: &BenchConfig) -> PolicyResult {
    let mut rng = SplitMix64::new(cfg.seed);
    let noise = policy.gradient_noise(cfg.steps_per_trajectory);
    let bias = policy.credit_bias(cfg.steps_per_trajectory);

    let mut curve = Vec::new();
    let mut episodes_to_target = None;
    let stride = (cfg.max_episodes / 40).max(1);

    for ep in 1..=cfg.max_episodes {
        let base = success_rate(policy, cfg.steps_per_trajectory, ep, cfg.ceiling);
        // Measurement jitter so the curve is not perfectly smooth.
        let jitter = (rng.next_f64() - 0.5) * 0.01;
        let sr = (base + jitter).clamp(0.0, 1.0);
        if episodes_to_target.is_none() && sr >= cfg.target_success_rate {
            episodes_to_target = Some(ep);
        }
        if ep % stride == 0 || ep == cfg.max_episodes {
            curve.push(CurvePoint { episode: ep, success_rate: sr });
        }
    }

    let final_success_rate = curve.last().map(|p| p.success_rate).unwrap_or(0.0);

    PolicyResult {
        policy: policy.label().into(),
        gradient_noise: noise,
        credit_bias: bias,
        episodes_to_target,
        final_success_rate,
        curve,
    }
}

pub fn run_bench(cfg: &BenchConfig) -> Result<BenchReport, LightningError> {
    if cfg.max_episodes == 0 {
        return Err(LightningError::InvalidConfig("max_episodes must be positive".into()));
    }
    if cfg.steps_per_trajectory == 0 {
        return Err(LightningError::InvalidConfig("steps_per_trajectory must be positive".into()));
    }
    if !(0.0..=1.0).contains(&cfg.target_success_rate) {
        return Err(LightningError::InvalidConfig("target_success_rate must be in [0,1]".into()));
    }
    if !(0.0..=1.0).contains(&cfg.ceiling) {
        return Err(LightningError::InvalidConfig("ceiling must be in [0,1]".into()));
    }
    let results = vec![
        evaluate(Policy::OutcomeOnly, cfg),
        evaluate(Policy::UniformCredit, cfg),
        evaluate(Policy::SegmentAdvantage, cfg),
    ];

    let mut notes = Vec::new();
    let fmt = |r: &PolicyResult| match r.episodes_to_target {
        Some(e) => format!("{} episodes", e),
        None => "never".to_string(),
    };
    if let (Some(out), Some(seg)) = (
        results.iter().find(|r| r.policy == "outcome-only"),
        results.iter().find(|r| r.policy == "segment-advantage"),
    ) {
        notes.push(format!(
            "segment advantages reach {:.0}% success in {} vs {} for outcome-only rewards; uniform credit plateaus at {:.1}% because its credit bias is {:.2}",
            seg.final_success_rate * 100.0,
            fmt(seg),
            fmt(out),
            results
                .iter()
                .find(|r| r.policy == "uniform-credit")
                .map(|r| r.final_success_rate * 100.0)
                .unwrap_or(0.0),
            results
                .iter()
                .find(|r| r.policy == "uniform-credit")
                .map(|r| r.credit_bias)
                .unwrap_or(0.0)
        ));
    }

    Ok(BenchReport {
        bounty: "hanzoai/agents#86 [Bounty #36] Learning from agent execution traces (Agent Lightning)".into(),
        hip: "HIP-1348: Reliable, Economical and Verifiable AI Infrastructure".into(),
        honesty_labels: vec![
            "no model is trained - the learning curve is a modelled function of estimator variance and credit bias".into(),
            "variance and bias constants are hand-set, not measured from a real RL run".into(),
            "trajectory step counts are synthetic".into(),
            "All figures are simulation outputs on the authoring host, NOT Hanzo expectations until reproduced with a real training loop.".into(),
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
    fn segment_advantage_has_lower_noise_than_outcome_only() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let seg = find(&report.results, "segment-advantage");
        let out = find(&report.results, "outcome-only");
        assert!(
            seg.gradient_noise < out.gradient_noise,
            "segment {} should be < outcome-only {}",
            seg.gradient_noise,
            out.gradient_noise
        );
        // Uniform credit is the lowest-variance estimator, but it pays for that
        // with a large systematic bias.
        let uni = find(&report.results, "uniform-credit");
        assert!(uni.gradient_noise <= seg.gradient_noise);
        assert!(uni.credit_bias > seg.credit_bias);
    }

    #[test]
    fn uniform_credit_is_biased() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let uni = find(&report.results, "uniform-credit");
        assert!(uni.credit_bias > 0.9, "got {}", uni.credit_bias);
        assert_eq!(find(&report.results, "outcome-only").credit_bias, 0.0);
    }

    #[test]
    fn segment_advantage_reaches_target_first() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let seg = find(&report.results, "segment-advantage").episodes_to_target.unwrap_or(usize::MAX);
        let out = find(&report.results, "outcome-only").episodes_to_target.unwrap_or(usize::MAX);
        assert!(seg <= out, "segment {} should be <= outcome {}", seg, out);
    }

    #[test]
    fn curves_are_monotone_enough() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        for r in &report.results {
            let first = r.curve.first().unwrap().success_rate;
            let last = r.curve.last().unwrap().success_rate;
            assert!(last > first, "{} did not improve ({:.3} -> {:.3})", r.policy, first, last);
        }
    }

    #[test]
    fn success_rate_respects_ceiling() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        for r in &report.results {
            for p in &r.curve {
                assert!(p.success_rate <= 1.0);
            }
            assert!(r.final_success_rate <= 0.82 + 0.01);
        }
    }

    #[test]
    fn bad_config_rejected() {
        let mut c = BenchConfig::test_config();
        c.max_episodes = 0;
        assert!(run_bench(&c).is_err());
        let mut c = BenchConfig::test_config();
        c.steps_per_trajectory = 0;
        assert!(run_bench(&c).is_err());
        let mut c = BenchConfig::test_config();
        c.ceiling = 2.0;
        assert!(run_bench(&c).is_err());
    }
}
