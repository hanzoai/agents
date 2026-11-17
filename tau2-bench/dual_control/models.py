from dataclasses import dataclass
from enum import Enum
from typing import Optional

class ActionSource(Enum):
    HUMAN = "human"
    AGENT = "agent"

@dataclass
class BusinessState:
    id: str
    status: str
    version: int
    last_modified_by: ActionSource

class StateConflictError(Exception):
    pass
