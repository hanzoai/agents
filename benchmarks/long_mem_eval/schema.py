from pydantic import BaseModel, Field
from typing import List, Optional, Literal

class MemoryEvent(BaseModel):
    timestamp: int
    content: str
    session_id: str
    metadata: Optional[dict] = {}

class EvalTask(BaseModel):
    id: str
    category: Literal["update", "contradiction", "temporal", "abstention", "cross_session"]
    events: List[MemoryEvent]
    query: str
    expected_answer: str
    negative_constraints: Optional[List[str]] = []

class EvalResult(BaseModel):
    task_id: str
    passed: bool
    model_output: str
    latency_ms: float
    retrieved_context_ids: List[int]
