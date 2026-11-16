from enum import Enum
from typing import Dict, Any, Optional, List
from pydantic import BaseModel, Field
import uuid
import datetime

class MemoryAction(str, Enum):
    ADD = "ADD"
    UPDATE = "UPDATE"
    NOOP = "NOOP"
    REVIEW = "REVIEW"

class Fact(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    content: str
    fact_type: str  # e.g., "preference", "entity_relation", "event"
    source_id: str  # Stable identity (Git commit, Agent event, etc.)
    scope: str      # tenant_id or session_id
    confidence: float = 1.0
    lifecycle: str = "active" # active, archived, deprecated
    created_at: str = Field(default_factory=lambda: datetime.datetime.utcnow().isoformat())
    updated_at: str = Field(default_factory=lambda: datetime.datetime.utcnow().isoformat())
    metadata: Dict[str, Any] = {}

class KaiDecision(BaseModel):
    action: MemoryAction
    target_id: Optional[str] = None
    reasoning: str
    refined_content: Optional[str] = None
