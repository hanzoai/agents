# Context Manager - Evidence Robustness Addendum

## Overview
Following HIP-1346, the `context-manager` agent must now prioritize evidence placement to mitigate "Lost in the Middle" phenomena.

## Evidence Ordering Strategy
1. **Primary Evidence**: Should be placed at the start or end of the context window.
2. **Secondary/Noise**: Should be packed in the middle.
3. **Token Optimization**: Use `context-bench` to prune irrelevant documents before injection to maintain cost-efficiency.

## Compliance
- All tool outputs must pass the `lost_in_the_middle.py` robustness check before being passed to `visor`.
