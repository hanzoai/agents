// SPDX-License-Identifier: Apache-2.0
//! [Bounty #39] (`hanzoai/agents#88`): **Dual-control business-agent benchmark** (tau2-bench).
//!
//! Clean-room simulation harness for business workflows where the agent and the
//! user share control of the same environment: either side can act, and some
//! actions are only legitimate when the user performs them.
//!
//! # Design
//!
//! Each task is a sequence of steps. Steps are one of:
//! - `AgentSafe`: the agent may perform it.
//! - `AgentRisky`: the agent *can* perform it but must confirm first.
//! - `UserOnly`: only the user has the authority (identity check, payment
//!   authorisation, policy exception).
//!
//! - `AgentOnly`: act on everything the agent can reach. Fast, violates
//!   authority boundaries.
//! - `ConfirmEveryStep`: ask before every action. Safe, unusable in practice.
//! - `DualControl`: hand off `UserOnly` steps, confirm `AgentRisky`, act
//!   autonomously on `AgentSafe`.
//!
//! Metrics: task success, authority violations, handoffs per task, and turns.
//!
//! # Honest boundaries
//! - No user simulator and no real business system; step types are synthetic.
//! - The probability a user grants a handoff and the cost of a violation are
//!   hand-set.
//! - All figures are simulation outputs on the authoring host, NOT Hanzo
//!   expectations until reproduced against a real tau2-bench environment.

use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum Tau2Error {
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
pub enum StepKind {
    AgentSafe,
    AgentRisky,
    UserOnly,
}

impl StepKind {
    pub fn name(&self) -> &'static str {
        match self {
            StepKind::AgentSafe => "agent-safe",
            StepKind::AgentRisky => "agent-risky",
            StepKind::UserOnly => "user-only",
        }
    }
}

#[derive(Debug, Clone)]
pub struct Task {
    pub id: u32,
    pub steps: Vec<StepKind>,
}

pub fn generate_tasks(seed: u64, n_tasks: usize, steps_per_task: usize, user_only_share: f64) -> Vec<Task> {
    let mut rng = SplitMix64::new(seed);
    (0..n_tasks)
        .map(|i| {
            let steps = (0..steps_per_task)
                .map(|_| {
                    let r = rng.next_f64();
                    if r < user_only_share {
                        StepKind::UserOnly
                    } else if r < user_only_share + 0.25 {
                        StepKind::AgentRisky
                    } else {
                        StepKind::AgentSafe
                    }
                })
                .collect();
            Task { id: i as u32, steps }
        })
        .collect()
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Policy {
    AgentOnly,
    ConfirmEveryStep,
    DualControl,
}

impl Policy {
    pub fn label(&self) -> &'static str {
        match self {
            Policy::AgentOnly => "agent-only",
            Policy::ConfirmEveryStep => "confirm-every-step",
            Policy::DualControl => "dual-control",
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PolicyResult {
    pub policy: String,
    pub tasks: usize,
    pub task_success_rate: f64,
    pub authority_violations: usize,
    pub violations_per_task: f64,
    pub handoffs_per_task: f64,
    pub turns_per_task: f64,
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
    pub steps_per_task: usize,
    /// Fraction of steps only the user may perform.
    pub user_only_share: f64,
    /// Probability the user cooperates when asked.
    pub user_cooperation: f64,
    /// Probability an autonomous risky action causes an incident.
    pub risky_incident_rate: f64,
}

impl Default for BenchConfig {
    fn default() -> Self {
        Self {
            seed: 0x7A02_BE11,
            n_tasks: 2_000,
            steps_per_task: 12,
            user_only_share: 0.22,
            user_cooperation: 0.93,
            risky_incident_rate: 0.34,
        }
    }
}
impl BenchConfig {
    pub fn test_config() -> Self {
        Self {
            seed: 7,
            n_tasks: 300,
            steps_per_task: 12,
            user_only_share: 0.22,
            user_cooperation: 0.93,
            risky_incident_rate: 0.34,
        }
    }
}

pub fn evaluate(policy: Policy, tasks: &[Task], cfg: &BenchConfig) -> PolicyResult {
    let mut rng = SplitMix64::new(cfg.seed ^ 0x7A02);
    let mut successes = 0usize;
    let mut violations = 0usize;
    let mut handoffs = 0usize;
    let mut turns = 0usize;

    for t in tasks {
        let mut ok = true;
        let mut task_handoffs = 0usize;
        let mut task_turns = 0usize;

        for s in &t.steps {
            task_turns += 1;
            match policy {
                Policy::AgentOnly => {
                    match s {
                        StepKind::AgentSafe => {}
                        StepKind::AgentRisky => {
                            // Acts without confirming: sometimes an incident.
                            if rng.next_f64() < cfg.risky_incident_rate {
                                violations += 1;
                                ok = false;
                            }
                        }
                        StepKind::UserOnly => {
                            // Performs an action it has no authority for.
                            violations += 1;
                            ok = false;
                        }
                    }
                }
                Policy::ConfirmEveryStep => {
                    // Asks the user about every single step.
                    task_handoffs += 1;
                    task_turns += 1;
                    if rng.next_f64() >= cfg.user_cooperation {
                        ok = false;
                    }
                }
                Policy::DualControl => match s {
                    StepKind::AgentSafe => {}
                    StepKind::AgentRisky => {
                        task_handoffs += 1;
                        task_turns += 1;
                        if rng.next_f64() >= cfg.user_cooperation {
                            // User declined; skip the risky action.
                            continue;
                        }
                        if rng.next_f64() < cfg.risky_incident_rate * 0.15 {
                            ok = false;
                        }
                    }
                    StepKind::UserOnly => {
                        task_handoffs += 1;
                        task_turns += 1;
                        if rng.next_f64() >= cfg.user_cooperation {
                            ok = false;
                        }
                    }
                },
            }
        }

        if ok {
            successes += 1;
        }
        handoffs += task_handoffs;
        turns += task_turns;
    }

    let n = tasks.len().max(1) as f64;
    PolicyResult {
        policy: policy.label().into(),
        tasks: tasks.len(),
        task_success_rate: successes as f64 / n,
        authority_violations: violations,
        violations_per_task: violations as f64 / n,
        handoffs_per_task: handoffs as f64 / n,
        turns_per_task: turns as f64 / n,
    }
}

pub fn run_bench(cfg: &BenchConfig) -> Result<BenchReport, Tau2Error> {
    if cfg.n_tasks == 0 {
        return Err(Tau2Error::InvalidConfig("n_tasks must be positive".into()));
    }
    if cfg.steps_per_task == 0 {
        return Err(Tau2Error::InvalidConfig("steps_per_task must be positive".into()));
    }
    if !(0.0..=1.0).contains(&cfg.user_only_share) {
        return Err(Tau2Error::InvalidConfig("user_only_share must be in [0,1]".into()));
    }
    if !(0.0..=1.0).contains(&cfg.user_cooperation) {
        return Err(Tau2Error::InvalidConfig("user_cooperation must be in [0,1]".into()));
    }
    if !(0.0..=1.0).contains(&cfg.risky_incident_rate) {
        return Err(Tau2Error::InvalidConfig("risky_incident_rate must be in [0,1]".into()));
    }
    let tasks = generate_tasks(cfg.seed, cfg.n_tasks, cfg.steps_per_task, cfg.user_only_share);
    let results = vec![
        evaluate(Policy::AgentOnly, &tasks, cfg),
        evaluate(Policy::ConfirmEveryStep, &tasks, cfg),
        evaluate(Policy::DualControl, &tasks, cfg),
    ];

    let mut notes = Vec::new();
    if let (Some(only), Some(dual)) = (
        results.iter().find(|r| r.policy == "agent-only"),
        results.iter().find(|r| r.policy == "dual-control"),
    ) {
        notes.push(format!(
            "dual control cuts authority violations from {:.2} to {:.2} per task while keeping {:.1}% task success (agent-only: {:.1}%), at {:.1} handoffs and {:.1} turns per task",
            only.violations_per_task,
            dual.violations_per_task,
            dual.task_success_rate * 100.0,
            only.task_success_rate * 100.0,
            dual.handoffs_per_task,
            dual.turns_per_task
        ));
    }

    Ok(BenchReport {
        bounty: "hanzoai/agents#88 [Bounty #39] Dual-control business-agent benchmark (tau2-bench)".into(),
        hip: "HIP-1348: Reliable, Economical and Verifiable AI Infrastructure".into(),
        honesty_labels: vec![
            "no user simulator and no real business system - step types are synthetic".into(),
            "user cooperation and incident probabilities are hand-set".into(),
            "the authority model is binary (user-only vs agent-allowed); real policies are richer".into(),
            "All figures are simulation outputs on the authoring host, NOT Hanzo expectations until reproduced against a real tau2-bench environment.".into(),
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
    fn dual_control_reduces_violations() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let only = find(&report.results, "agent-only");
        let dual = find(&report.results, "dual-control");
        assert!(
            dual.violations_per_task < only.violations_per_task,
            "dual {:.3} should be < agent-only {:.3}",
            dual.violations_per_task,
            only.violations_per_task
        );
    }

    #[test]
    fn dual_control_beats_confirm_every_step_on_success() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        let confirm = find(&report.results, "confirm-every-step");
        let dual = find(&report.results, "dual-control");
        assert!(dual.task_success_rate >= confirm.task_success_rate);
        assert!(dual.handoffs_per_task < confirm.handoffs_per_task);
    }

    #[test]
    fn agent_only_has_no_handoffs() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        assert_eq!(find(&report.results, "agent-only").handoffs_per_task, 0.0);
    }

    #[test]
    fn confirm_every_step_hands_off_on_every_step() {
        let cfg = BenchConfig::test_config();
        let report = run_bench(&cfg).unwrap();
        let confirm = find(&report.results, "confirm-every-step");
        assert!((confirm.handoffs_per_task - cfg.steps_per_task as f64).abs() < 1e-9);
    }

    #[test]
    fn dual_control_hands_off_less_than_every_step() {
        let cfg = BenchConfig::test_config();
        let report = run_bench(&cfg).unwrap();
        let dual = find(&report.results, "dual-control");
        assert!(dual.handoffs_per_task < cfg.steps_per_task as f64);
        assert!(dual.handoffs_per_task > 0.0);
    }

    #[test]
    fn bad_config_rejected() {
        let mut c = BenchConfig::test_config();
        c.n_tasks = 0;
        assert!(run_bench(&c).is_err());
        let mut c = BenchConfig::test_config();
        c.user_cooperation = 1.5;
        assert!(run_bench(&c).is_err());
        let mut c = BenchConfig::test_config();
        c.risky_incident_rate = -0.1;
        assert!(run_bench(&c).is_err());
    }
}
