from pydantic import BaseModel, Field
from typing import List, Optional, Dict, Any

class TraceStep(BaseModel):
    step_id: str
    thought: str
    action: str
    observation: str
    reward: float
    metadata: Dict[str, Any] = {}

class Trajectory(BaseModel):
    trajectory_id: str
    agent_id: str
    steps: List[TraceStep]
    success: bool
    total_reward: float
