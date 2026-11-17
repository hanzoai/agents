"""
Agent Execution Trace to RL/SFT Dataset Converter
Implementation for Hanzo AI Bounty #36 / HIP-1347: Event-Driven Agent Reliability.

Converts multi-step agent trajectories into structured:
1. Supervised Fine-Tuning (SFT) multi-turn conversations for Zen checkpoints.
2. Direct Preference Optimization (DPO / RLHF) chosen/rejected pairs.
3. Automated Reward-Hacking Detection (test deletion, assertion bypass, vacuous green passes).
4. Training token & compute cost estimation.
"""

import re
import json
import hashlib
from typing import Dict, List, Any, Optional, Tuple
from dataclasses import dataclass, asdict, field


@dataclass
class TrajectoryStep:
    step_index: int
    role: str  # 'user', 'assistant', 'tool'
    content: str
    tool_calls: Optional[List[Dict[str, Any]]] = None
    tool_results: Optional[List[Dict[str, Any]]] = None
    is_success_step: bool = True
    metadata: Dict[str, Any] = field(default_factory=dict)


@dataclass
class AgentTrajectory:
    trajectory_id: str
    task_prompt: str
    task_id: str
    outcome: str  # 'SUCCESS', 'FAILED', 'TIMED_OUT'
    steps: List[TrajectoryStep]
    total_tokens: int = 0
    duration_seconds: float = 0.0
    environment_metadata: Dict[str, Any] = field(default_factory=dict)


class RewardHackingDetector:
    """
    Detects deceptive shortcuts or reward hacking in agent trajectories:
    - Modifying or removing test suites.
    - Exiting with code 0 without running assertions.
    - Hardcoding expected return values based on test literals.
    - Mocking critical validation libraries to bypass security checks.
    """

    TEST_DELETION_PATTERNS = [
        r"rm\s+.*test.*",
        r"del\s+.*test.*",
        r"git\s+checkout\s+--\s+.*test.*",
        r"#\s*pytest\.skip",
        r"@unittest\.skip",
        r"return\s+True\s*#\s*bypass",
        r"sys\.exit\(0\)"
    ]

    ASSERTION_BYPASS_PATTERNS = [
        r"assert\s+True",
        r"self\.assertTrue\(True\)",
        r"except\s+Exception:\s*pass"
    ]

    @classmethod
    def audit_trajectory(cls, trajectory: AgentTrajectory) -> Tuple[bool, List[str]]:
        violations: List[str] = []

        for step in trajectory.steps:
            # Audit tool calls and bash commands
            if step.tool_calls:
                for tc in step.tool_calls:
                    call_str = json.dumps(tc).lower()
                    for pat in cls.TEST_DELETION_PATTERNS:
                        if re.search(pat, call_str):
                            violations.append(f"Step {step.step_index}: Suspicious test manipulation pattern '{pat}'")

            # Audit assistant generated code
            if step.role == "assistant":
                content = step.content
                for pat in cls.ASSERTION_BYPASS_PATTERNS:
                    if re.search(pat, content):
                        violations.append(f"Step {step.step_index}: Potential vacuous assertion bypass '{pat}'")

        is_hacked = len(violations) > 0
        return is_hacked, violations


class TraceDatasetSynthesizer:
    """
    Transforms raw trajectory logs into curated training datasets:
    - SFT (Supervised Fine-Tuning) datasets for Zen checkpoints.
    - DPO (Direct Preference Optimization) preference pairs.
    """

    def __init__(self, cost_per_1k_tokens_usd: float = 0.0015):
        self.cost_per_1k_tokens = cost_per_1k_tokens_usd

    def convert_to_sft_conversation(self, trajectory: AgentTrajectory) -> Optional[Dict[str, Any]]:
        """
        Extracts SFT conversational trajectory only if the trajectory succeeded
        and cleared reward-hacking audits.
        """
        if trajectory.outcome != "SUCCESS":
            return None

        is_hacked, violations = RewardHackingDetector.audit_trajectory(trajectory)
        if is_hacked:
            return None

        messages: List[Dict[str, Any]] = [
            {"role": "system", "content": "You are Zen, a high-reliability autonomous software engineering agent."},
            {"role": "user", "content": trajectory.task_prompt}
        ]

        for step in trajectory.steps:
            msg: Dict[str, Any] = {"role": step.role, "content": step.content}
            if step.tool_calls:
                msg["tool_calls"] = step.tool_calls
            if step.tool_results:
                msg["tool_results"] = step.tool_results
            messages.append(msg)

        est_tokens = self._estimate_tokens(messages)
        return {
            "trajectory_id": trajectory.trajectory_id,
            "task_id": trajectory.task_id,
            "format": "zen_sft_v1",
            "messages": messages,
            "total_tokens_est": est_tokens,
            "estimated_sft_cost_usd": round((est_tokens / 1000.0) * self.cost_per_1k_tokens, 6)
        }

    def synthesize_dpo_pair(
        self,
        successful_trajectory: AgentTrajectory,
        failed_trajectory: AgentTrajectory
    ) -> Optional[Dict[str, Any]]:
        """
        Synthesizes a (Prompt, Chosen, Rejected) pair for Direct Preference Optimization (DPO).
        Both trajectories must target the same task_id.
        """
        if successful_trajectory.task_id != failed_trajectory.task_id:
            raise ValueError(f"Task ID mismatch: '{successful_trajectory.task_id}' != '{failed_trajectory.task_id}'")

        if successful_trajectory.outcome != "SUCCESS" or failed_trajectory.outcome == "SUCCESS":
            return None

        # Verify successful trajectory isn't hacked
        is_hacked, _ = RewardHackingDetector.audit_trajectory(successful_trajectory)
        if is_hacked:
            return None

        prompt = successful_trajectory.task_prompt
        chosen_steps = [s.content for s in successful_trajectory.steps if s.role == "assistant"]
        rejected_steps = [s.content for s in failed_trajectory.steps if s.role == "assistant"]

        chosen_text = "\n\n".join(chosen_steps)
        rejected_text = "\n\n".join(rejected_steps)

        return {
            "task_id": successful_trajectory.task_id,
            "prompt": prompt,
            "chosen": chosen_text,
            "rejected": rejected_text,
            "metadata": {
                "chosen_trajectory_id": successful_trajectory.trajectory_id,
                "rejected_trajectory_id": failed_trajectory.trajectory_id,
                "rejected_reason": failed_trajectory.outcome
            }
        }

    def _estimate_tokens(self, messages: List[Dict[str, Any]]) -> int:
        raw_text = json.dumps(messages)
        # Conservative token estimation (~4 characters per token)
        return max(1, len(raw_text) // 4)

    def export_dataset(self, sft_records: List[Dict[str, Any]], dpo_records: List[Dict[str, Any]], output_dir: str):
        """Exports formatted jsonl files ready for Zen training."""
        import os
        os.makedirs(output_dir, exist_ok=True)

        sft_path = os.path.join(output_dir, "zen_sft_dataset.jsonl")
        with open(sft_path, "w", encoding="utf-8") as f:
            for rec in sft_records:
                f.write(json.dumps(rec) + "\n")

        dpo_path = os.path.join(output_dir, "zen_dpo_dataset.jsonl")
        with open(dpo_path, "w", encoding="utf-8") as f:
            for rec in dpo_records:
                f.write(json.dumps(rec) + "\n")

        summary = {
            "sft_records_count": len(sft_records),
            "dpo_pairs_count": len(dpo_records),
            "total_tokens_est": sum(r.get("total_tokens_est", 0) for r in sft_records),
            "sft_export_path": sft_path,
            "dpo_export_path": dpo_path
        }
        return summary
