from enum import Enum
from typing import Any, Dict, Optional, List
from pydantic import BaseModel, Field
import uuid
import time

class MemoryTier(str, Enum):
    EPISODIC = "episodic"
    SEMANTIC = "semantic"
    PROCEDURAL = "procedural"

class MemoryRecord(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    version: int = 1
    tenant_id: str
    tier: MemoryTier
    content: Any
    metadata: Dict[str, Any] = Field(default_factory=dict)
    timestamp: float = Field(default_factory=time.time)
    parent_id: Optional[str] = None

class EpisodicRecord(MemoryRecord):
    tier: MemoryTier = MemoryTier.EPISODIC
    event_type: str  # observation, action, result

class SemanticRecord(MemoryRecord):
    tier: MemoryTier = MemoryTier.SEMANTIC
    entity_type: str # fact, relationship, decision

class ProceduralRecord(MemoryRecord):
    tier: MemoryTier = MemoryTier.PROCEDURAL
    name: str
    success_count: int = 0
    failure_count: int = 0
