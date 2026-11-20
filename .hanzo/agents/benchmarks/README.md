# Benchmarks for Evidence-Grounded Retrieval

This directory contains tools to validate the robustness of agent tool outputs against the "Lost in the Middle" research findings.

## Running the Benchmark
1. Ensure your environment is configured with `hanzoai/s3` access.
2. Execute: `python3 .hanzo/agents/benchmarks/lost_in_the_middle.py`
3. Review `results.json` to verify if the agent maintains >80% accuracy regardless of evidence position.

## HIP-1346 Status
- [x] Baseline established.
- [ ] Integration with `visor` pipeline (Pending).
