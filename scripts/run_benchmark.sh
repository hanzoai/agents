#!/bin/bash
# Reproducible baseline runner for SWE-bench Pro V2
set -e

echo "Starting SWE-bench Pro V2 evaluation..."

# Test 1: VFS Performance
python3 -m eval.swe_bench_runner --fs vfs --output results_vfs.json

# Test 2: NVMe Performance
python3 -m eval.swe_bench_runner --fs nvme --output results_nvme.json

echo "Benchmark complete. Results saved to results_*.json"
