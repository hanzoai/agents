// SPDX-License-Identifier: Apache-2.0
use agents_longmem::{BenchConfig, run_bench, write_results_json};
use std::path::PathBuf;
fn main() {
    let out = std::env::args().nth(1).map(PathBuf::from).unwrap_or_else(|| PathBuf::from("results/longmem-results.json"));
    let cfg = BenchConfig::default();
    eprintln!("agents-longmem bench");
    let report = run_bench(&cfg).expect("failed");
    write_results_json(&report, &out).expect("failed");
    eprintln!("wrote {}", out.display());
}