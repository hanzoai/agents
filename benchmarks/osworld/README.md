# OSWorld Benchmark for Visor

This module provides the tooling to run OSWorld tasks within the `hanzoai/visor` sandbox environment.

## Metrics tracked:
- **Success Rate**: Task completion based on VFS state verification.
- **Cold-start**: Time from `visor` initialization to first agent action.
- **Footprint**: CPU/RAM usage sampled at 1Hz.
- **Cleanup**: Verification of sandbox state post-teardown.

## How to run:
1. Ensure `visor` is configured with the canonical stack (`vfs` -> `s3`).
2. Run `python3 -m benchmarks.osworld.run --task-id <id>`.
