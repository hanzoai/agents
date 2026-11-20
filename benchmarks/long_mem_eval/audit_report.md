# Research Audit: LongMemEval

## 1. Primary Reference
- **Paper**: [LongMemEval: Benchmarking Long-term Memory in Agents](https://proceedings.iclr.cc/paper_files/paper/2025/hash/d813d324dbf0598bbdc9c8e79740ed01-Abstract-Conference.html) (ICLR 2025).
- **Code**: [xiaowu0162/LongMemEval](https://github.com/xiaowu0162/LongMemEval).

## 2. Design Summary
LongMemEval shifts focus from simple "long context" to "long-term memory" dynamics. It evaluates how agents manage a persistent knowledge base that changes over time. Unlike static RAG benchmarks, it tests:
- **Point-in-time correctness**: Information valid at $T_1$ may be invalid at $T_2$.
- **Conflict Resolution**: Handling contradictory statements across sessions.
- **Negative Constraints**: Testing if the agent knows what it *doesn't* know.

## 3. License & Patents
- **License**: The original code is licensed under **Apache-2.0**. This implementation is a clean-room adaptation designed to integrate with Hanzo's proprietary `vfs` and `visor` architecture.
- **Compatibility**: Fully compatible with Hanzo's production requirements. No known patent restrictions.

## 4. HIP-1346 Alignment
This deliverable provides the "Evidence-Grounded Retrieval" evaluation layer required by HIP-1346, ensuring that memory retrieval is not just relevant but temporally accurate.
