from typing import List, Dict, Any
from .models import Fact, MemoryAction
from .kai_engine import KaiEngine
from .storage import FactStore, VectorIndex

class MemoryManager:
    def __init__(self):
        self.store = FactStore()
        self.vector_index = VectorIndex()
        self.kai = KaiEngine()

    def process_event(self, text: str, metadata: Dict[str, Any]):
        """
        Event-driven writer that extracts facts and reconciles them.
        """
        # 1. Extract facts (Simulated LLM call)
        extracted_facts = self._extract_facts_from_text(text, metadata)
        
        for raw_fact in extracted_facts:
            # 2. Retrieve relevant context for Kai
            scope = metadata.get("scope", "default")
            existing_context = self.store.get_by_scope(scope)
            
            # 3. Kai Decision
            decision = self.kai.decide(raw_fact, existing_context)
            
            # 4. Execute Decision
            if decision.action == MemoryAction.ADD:
                self.store.add(raw_fact)
            elif decision.action == MemoryAction.UPDATE:
                self.store.update(decision.target_id, decision.refined_content)
            elif decision.action == MemoryAction.REVIEW:
                # Log for human or supervisor agent review
                print(f"CONFLICT: {decision.reasoning}")
            
        # 5. Refresh derived indexes
        self.vector_index.rebuild(self.store.get_by_scope(scope))

    def _extract_facts_from_text(self, text: str, metadata: Dict[str, Any]) -> List[Fact]:
        # In production, this uses an LLM to parse text into atomic Fact objects.
        # Mocking extraction:
        return [
            Fact(
                content=text[:100], 
                fact_type="observation", 
                source_id=metadata.get("source_id", "unknown"),
                scope=metadata.get("scope", "default"),
                metadata=metadata
            )
        ]

    def retrieve(self, query: str, scope: str) -> List[Fact]:
        # Apply tenant isolation before returning hits
        all_facts = self.store.get_by_scope(scope)
        # Further filtering logic here...
        return all_facts
