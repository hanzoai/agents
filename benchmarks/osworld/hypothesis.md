# Falsifiable Hypothesis (HIP-1347)

**Hypothesis**: The `hanzoai/visor` sandbox, when utilizing the canonical `vfs -> s3` path, will maintain a cold-start time of < 2.5s and a resource overhead of < 15% CPU/RAM compared to a native execution environment, while ensuring 100% cleanup of ephemeral state.

**Metrics**:
- Target Cold-start: < 2.5s
- Target Success Rate: > 80% on OSWorld subset
- Cleanup: 0 leaked files in VFS post-teardown.
