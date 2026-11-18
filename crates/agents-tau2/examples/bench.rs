// SPDX-License-Identifier: Apache-2.0
//! Reproducible runner.
use agents_tau2::{BenchConfig, run_bench, write_results_json};
use std::path::PathBuf;

fn main() {
    let out = std::env::args()
        .nth(1)
        .map(PathBuf::from)
        .unwrap_or_else(|| PathBuf::from("results.json"));
    let cfg = BenchConfig::default();
    eprintln!("running agents-tau2");
    let report = run_bench(&cfg).expect("benchmark failed");
    write_results_json(&report, &out).expect("failed to write results");
    eprintln!("wrote {}", out.display());
}
