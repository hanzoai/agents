import json
from typing import List, Optional
from .models import Fact, KaiDecision, MemoryAction

class KaiEngine:
    """
    Kai: A small structured-decision model logic for memory reconciliation.
    In a production environment, this would call a quantized model (like Phi-3 or Llama-3-8B)
    with a specific system prompt.
    """
    
    def __init__(self, llm_client=None):
        self.llm_client = llm_client

    def decide(self, new_fact: Fact, existing_memories: List[Fact]) -> KaiDecision:
        if not existing_memories:
            return KaiDecision(action=MemoryAction.ADD, reasoning="No existing context found.")

        # Logic for reconciliation
        for existing in existing_memories:
            # 1. Exact or high-semantic match
            if self._is_duplicate(new_fact, existing):
                return KaiDecision(action=MemoryAction.NOOP, target_id=existing.id, reasoning="Fact already exists.")
            
            # 2. Contradiction or Update
            if self._is_update(new_fact, existing):
                return KaiDecision(
                    action=MemoryAction.UPDATE, 
                    target_id=existing.id, 
                    reasoning="New information refines existing fact.",
                    refined_content=new_fact.content
                )

            # 3. Conflict requiring review
            if self._is_conflict(new_fact, existing):
                return KaiDecision(action=MemoryAction.REVIEW, target_id=existing.id, reasoning="Direct contradiction detected.")

        return KaiDecision(action=MemoryAction.ADD, reasoning="New unique information.")

    def _is_duplicate(self, new: Fact, old: Fact) -> bool:
        # Placeholder for semantic similarity check > 0.95
        return new.content.lower().strip() == old.content.lower().strip()

    def _is_update(self, new: Fact, old: Fact) -> bool:
        # Logic to detect if 'new' is a more recent or detailed version of 'old'
        # e.g., "User lives in NY" -> "User lives in Brooklyn, NY"
        return old.fact_type == new.fact_type and new.content.startswith(old.content)

    def _is_conflict(self, new: Fact, old: Fact) -> bool:
        # Logic to detect contradictions
        # e.g., "User is 25" vs "User is 30"
        return False # Simplified for implementation
