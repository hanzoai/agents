import uuid
from .models import BusinessState, ActionSource, StateConflictError

class DualControlEngine:
    def __init__(self):
        self.store = {} # Mocking vfs/s3 layer

    def mutate(self, entity_id: str, new_status: str, source: ActionSource):
        current = self.store.get(entity_id, BusinessState(entity_id, "INIT", 0, ActionSource.HUMAN))
        
        # Permission & Concurrency Check
        if source == ActionSource.AGENT and current.status == "LOCKED_BY_HUMAN":
            raise StateConflictError("Agent cannot mutate state locked by human.")
            
        # Idempotent Update
        current.status = new_status
        current.version += 1
        current.last_modified_by = source
        self.store[entity_id] = current
        return current
