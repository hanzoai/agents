from typing import List, Dict, Optional
from .models import Fact

class FactStore:
    """
    Source of truth for atomic facts. 
    Indexes (Vector, Lexical) are DERIVED from this store.
    """
    def __init__(self):
        self._storage: Dict[str, Fact] = {}

    def add(self, fact: Fact):
        self._storage[fact.id] = fact

    def update(self, fact_id: str, new_content: str):
        if fact_id in self._storage:
            self._storage[fact_id].content = new_content
            self._storage[fact_id].updated_at = "now" # Use real timestamp

    def get_by_scope(self, scope: str) -> List[Fact]:
        return [f for f in self._storage.values() if f.scope == scope]

    def delete(self, fact_id: str):
        if fact_id in self._storage:
            del self._storage[fact_id]

class DerivedIndex:
    """Base class for rebuildable indexes."""
    def rebuild(self, facts: List[Fact]):
        raise NotImplementedError("Indexes must implement rebuild logic.")

class VectorIndex(DerivedIndex):
    def __init__(self):
        self.index = {} # Placeholder for actual vector DB (Chroma/Pinecone)

    def rebuild(self, facts: List[Fact]):
        self.index = {f.id: "vector_representation" for f in facts}
        print(f"Rebuilt vector index with {len(facts)} facts.")

    def search(self, query_vector: List[float], limit: int = 5) -> List[str]:
        # Return fact IDs
        return list(self.index.keys())[:limit]
