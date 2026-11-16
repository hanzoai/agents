from typing import Any, Dict, List, Optional
from .types import MemoryTier, MemoryRecord, EpisodicRecord, SemanticRecord, ProceduralRecord
from .storage import SQLiteMemoryStorage
import time

class MemoryManager:
    """
    Unified API for Episodic, Semantic, and Procedural Memory.
    """
    def __init__(self, storage: Optional[SQLiteMemoryStorage] = None):
        self.storage = storage or SQLiteMemoryStorage()

    def remember(self, 
                 tier: MemoryTier, 
                 tenant_id: str, 
                 content: Any, 
                 metadata: Dict[str, Any] = None,
                 **kwargs) -> str:
        
        if tier == MemoryTier.EPISODIC:
            record = EpisodicRecord(
                tenant_id=tenant_id,
                content=content,
                metadata=metadata or {},
                event_type=kwargs.get("event_type", "observation")
            )
        elif tier == MemoryTier.SEMANTIC:
            record = SemanticRecord(
                tenant_id=tenant_id,
                content=content,
                metadata=metadata or {},
                entity_type=kwargs.get("entity_type", "fact")
            )
        elif tier == MemoryTier.PROCEDURAL:
            record = ProceduralRecord(
                tenant_id=tenant_id,
                content=content,
                metadata=metadata or {},
                name=kwargs.get("name", "unnamed_procedure")
            )
        else:
            raise ValueError(f"Unsupported memory tier: {tier}")

        self.storage.store(record)
        return record.id

    def recall(self, tenant_id: str, tier: MemoryTier, query: str = None, limit: int = 10) -> List[Dict[str, Any]]:
        # In a full implementation, 'query' would trigger a vector search on derived indexes.
        # Here we implement the base retrieval from durable storage.
        records = self.storage.retrieve(tenant_id, tier, limit)
        return records

    def compact(self, tenant_id: str, tier: MemoryTier):
        """
        Triggers compaction logic specific to the tier.
        """
        if tier == MemoryTier.EPISODIC:
            self._compact_episodic(tenant_id)
        elif tier == MemoryTier.SEMANTIC:
            self._compact_semantic(tenant_id)

    def _compact_episodic(self, tenant_id: str):
        # Logic: Fetch old episodic logs, summarize them, store summary in Semantic, delete logs.
        logs = self.storage.retrieve(tenant_id, MemoryTier.EPISODIC, limit=50)
        if len(logs) < 10: return
        
        summary_content = f"Summary of {len(logs)} events ending at {logs[0]['timestamp']}"
        self.remember(
            tier=MemoryTier.SEMANTIC,
            tenant_id=tenant_id,
            content={"summary": summary_content, "source_count": len(logs)},
            entity_type="event_summary"
        )
        # In production, delete or archive the compacted episodic records here.

    def _compact_semantic(self, tenant_id: str):
        # Logic: Deduplicate facts based on content hashes or LLM evaluation.
        pass
