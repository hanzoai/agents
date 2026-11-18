// SPDX-License-Identifier: Apache-2.0
//! [Bounty #29] (`hanzoai/agents#83`): **Evidence-ordering robustness** (LostInTheMiddle).
//!
//! Clean-room evaluation harness benchmarking evidence-position bias per
//! *Lost in the Middle* (Liu et al., 2023): measure retrieval accuracy when
//! relevant evidence appears at different positions (beginning/middle/end) in
//! the context window, using deterministic agent responses.
//!
//! # Design
//!
//! - Generate Q&A pairs with evidence at varying positions.
//! - Simulate agent reading order bias (first/last position advantage).
//! - Measure answer correctness vs evidence position.

use serde::{Deserialize, Serialize};
use thiserror::Error;

#[derive(Debug, Error)]
pub enum LIMError { #[error("invalid config: {0}")] InvalidConfig(String) }

#[derive(Debug, Clone)]
pub struct EvalCase {
    pub question: String,
    pub evidence_chunks: Vec<String>,
    pub correct_answer: String,
    pub evidence_position: usize, // which chunk index contains the key evidence
}

pub fn generate_cases(seed: u64, n: usize) -> Vec<EvalCase> {
    let mut rng = SplitMix64::new(seed);
    (0..n).map(|i| {
        let n_chunks = 3 + (rng.next_u64() as usize % 5); // 3-7 chunks
        let pos = rng.next_u64() as usize % n_chunks;
        let chunks: Vec<String> = (0..n_chunks)
            .map(|j| format!("chunk_{}_{}", j, rng.next_u64()))
            .collect();
        EvalCase {
            question: format!("question_{}", i),
            evidence_chunks: chunks,
            correct_answer: format!("answer_{}", i),
            evidence_position: pos,
        }
    }).collect()
}

#[derive(Debug, Clone)]
pub struct SplitMix64 { state: u64 }
impl SplitMix64 { pub fn new(s: u64) -> Self { Self { state: s } }
    pub fn next_u64(&mut self) -> u64 {
        self.state = self.state.wrapping_add(0x9E37_79B9_7F4A_7C15);
        let mut z = self.state;
        z = (z ^ (z >> 30)).wrapping_mul(0xBF58_476D_1CE4_E5B9);
        z = (z ^ (z >> 27)).wrapping_mul(0x94D0_49BB_1331_11EB);
        z ^ (z >> 31)
    }
}

/// Simulate agent recall: first/last chunk positions get boost (position bias).
fn simulate_recall(case: &EvalCase, bias_strength: f64) -> bool {
    let total = case.evidence_chunks.len();
    let pos = case.evidence_position;
    // First and last positions have recall boost
    let is_first = pos == 0;
    let is_last = pos == total - 1;
    let base_prob = 0.5;
    let boost = if is_first || is_last { bias_strength } else { 0.0 };
    // Deterministic: compare hash against threshold
    let mut h: u64 = pos as u64 * 2654435761u64 ^ case.correct_answer.len() as u64;
    h = h.wrapping_mul(0x9E37_79B9_7F4A_7C15);
    let threshold = ((base_prob + boost) * 100.0) as u64;
    (h % 100) < threshold
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct PositionStats {
    pub position: String,
    pub n_cases: usize,
    pub recall_rate: f32,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct RunResult {
    pub policy: String,
    pub n_cases: usize,
    pub overall_recall: f32,
    pub by_position: Vec<PositionStats>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BenchReport {
    pub bounty: String,
    pub hip: String,
    pub honesty_labels: Vec<String>,
    pub runs: Vec<RunResult>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct BenchConfig {
    pub seed: u64,
    pub n_cases: usize,
    pub bias_strength: f64, // how strong is the position bias
}
impl Default for BenchConfig { Self { seed: 0x21, n_cases: 200, bias_strength: 0.25 } }
impl BenchConfig {
    pub fn test_config() -> Self { Self { seed: 7, n_cases: 30, bias_strength: 0.3 } }
}

pub fn run_bench(cfg: &BenchConfig) -> Result<BenchReport, LIMError> {
    if cfg.n_cases == 0 { return Err(LIMError::InvalidConfig("n_cases must be positive".into())); }
    let cases = generate_cases(cfg.seed, cfg.n_cases);
    let mut runs: Vec<RunResult> = Vec::new();

    for bias in [0.0, cfg.bias_strength, cfg.bias_strength * 2.0] {
        let mut correct = 0usize;
        let mut by_pos: std::collections::HashMap<usize, (usize, usize)> = std::collections::HashMap::new();
        
        for case in &cases {
            let hit = simulate_recall(case, bias);
            if hit { correct += 1; }
            let entry = by_pos.entry(case.evidence_position).or_insert((0, 0));
            entry.1 += 1;
            if hit { entry.0 += 1; }
        }
        
        let recall = correct as f32 / cfg.n_cases as f32;
        let by_position: Vec<PositionStats> = by_pos.iter()
            .map(|&(pos, (hit, total))| PositionStats {
                position: format!("chunk_{}", pos),
                n_cases: total,
                recall_rate: hit as f32 / total as f32,
            }).collect();
        
        runs.push(RunResult {
            policy: format!("bias={:.2}", bias),
            n_cases: cfg.n_cases,
            overall_recall: recall,
            by_position,
        });
    }

    Ok(BenchReport {
        bounty: "hanzoai/agents#83 [Bounty #29] Evidence-ordering robustness (LostInTheMiddle)".into(),
        hip: "HIP-1346: Evidence-Grounded Retrieval and Decision Evaluation".into(),
        honesty_labels: vec![
            "simulated agent recall — no real LLM or agent execution".into(),
            "recall model uses deterministic position-bias heuristic, not real cognitive bias data".into(),
            "All figures are simulation outputs on the authoring host, NOT Hanzo expectations until reproduced with real agent traces.".into(),
        ],
        runs,
    })
}

pub fn write_results_json(r: &BenchReport, p: &std::path::Path) -> Result<(), Box<dyn std::error::Error>> {
    if let Some(parent) = p.parent() { std::fs::create_dir_all(parent)?; }
    std::fs::write(p, serde_json::to_string_pretty(r)?)?; Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn bench_runs_cleanly() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        assert_eq!(report.runs.len(), 3);
        for r in &report.runs {
            assert_eq!(r.n_cases, 30);
            assert!(r.overall_recall >= 0.0 && r.overall_recall <= 1.0);
        }
    }
    #[test]
    fn higher_bias_improves_recall() {
        let report = run_bench(&BenchConfig::test_config()).unwrap();
        // Last run should have highest bias and thus highest recall
        let recalls: Vec<f32> = report.runs.iter().map(|r| r.overall_recall).collect();
        assert!(recalls[2] >= recalls[0], "higher bias should improve recall: {:?})", recalls);
    }
}
