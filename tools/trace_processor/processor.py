import json
import os
import re
from typing import List, Dict
from .models import Trajectory, TraceStep

class ZenDatasetGenerator:
    """
    Converts raw agent trajectories into Zen-compatible SFT/RL datasets.
    Follows HIP-1347 decoupling of execution and optimization.
    """
    
    def __init__(self, pii_masking: bool = True):
        self.pii_masking = pii_masking
        # Basic PII patterns
        self.pii_patterns = [
            re.compile(r'sk-[a-zA-Z0-9]{32,}'), # OpenAI-like keys
            re.compile(r'AIza[0-9A-Za-z-_]{35}'), # Google-like keys
            re.compile(r'[\w\.-]+@[\w\.-]+\.\w+') # Emails
        ]

    def mask_content(self, text: str) -> str:
        if not self.pii_masking:
            return text
        for pattern in self.pii_patterns:
            text = pattern.sub("[MASKED]", text)
        return text

    def to_sft_format(self, trajectory: Trajectory) -> List[Dict]:
        """
        Converts a successful trajectory into a series of instruction-response pairs.
        """
        if not trajectory.success:
            return []

        dataset_entries = []
        context = ""
        
        for step in trajectory.steps:
            prompt = f"Context: {context}\nAction needed for task."
            # Zen format expects 'instruction' and 'output'
            entry = {
                "instruction": self.mask_content(prompt),
                "thought": self.mask_content(step.thought),
                "output": self.mask_content(step.action)
            }
            dataset_entries.append(entry)
            context += f"\nAction: {step.action}\nObservation: {step.observation}"
            
        return dataset_entries

    def to_dpo_format(self, success_traj: Trajectory, failed_traj: Trajectory) -> Dict:
        """
        Creates a preference pair for RLHF/DPO.
        """
        return {
            "prompt": self.mask_content(success_traj.steps[0].thought), # Simplified
            "chosen": self.mask_content(success_traj.steps[0].action),
            "rejected": self.mask_content(failed_traj.steps[0].action)
        }

def process_visor_logs(log_dir: str) -> List[Trajectory]:
    """
    Parses logs from the hanzoai/visor format.
    Expected path: .hanzo/traces/raw/*.json
    """
    trajectories = []
    if not os.path.exists(log_dir):
        return []

    for file_name in os.listdir(log_dir):
        if file_name.endswith(".json"):
            with open(os.path.join(log_dir, file_name), 'r') as f:
                data = json.load(f)
                # Mapping logic from Visor schema to Trajectory model
                steps = [TraceStep(**s) for s in data.get("steps", [])]
                trajectories.append(Trajectory(
                    trajectory_id=data.get("id"),
                    agent_id=data.get("agent_id"),
                    steps=steps,
                    success=data.get("success", False),
                    total_reward=data.get("total_reward", 0.0)
                ))
    return trajectories
