import sqlite3
import json
from typing import List, Optional, Dict, Any
from .types import MemoryRecord, MemoryTier

class SQLiteMemoryStorage:
    def __init__(self, db_path: str = "hanzo_memory.db"):
        self.db_path = db_path
        self._init_db()

    def _init_db(self):
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS records (
                    id TEXT PRIMARY KEY,
                    version INTEGER,
                    tenant_id TEXT,
                    tier TEXT,
                    content TEXT,
                    metadata TEXT,
                    timestamp REAL,
                    parent_id TEXT
                )
            """)
            conn.execute("CREATE INDEX IF NOT EXISTS idx_tenant_tier ON records(tenant_id, tier)")

    def store(self, record: MemoryRecord):
        with sqlite3.connect(self.db_path) as conn:
            conn.execute("""
                INSERT OR REPLACE INTO records (id, version, tenant_id, tier, content, metadata, timestamp, parent_id)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            """, (
                record.id, record.version, record.tenant_id, record.tier.value,
                json.dumps(record.content), json.dumps(record.metadata),
                record.timestamp, record.parent_id
            ))

    def retrieve(self, tenant_id: str, tier: MemoryTier, limit: int = 100) -> List[Dict[str, Any]]:
        with sqlite3.connect(self.db_path) as conn:
            conn.row_factory = sqlite3.Row
            cursor = conn.execute(
                "SELECT * FROM records WHERE tenant_id = ? AND tier = ? ORDER BY timestamp DESC LIMIT ?",
                (tenant_id, tier.value, limit)
            )
            return [dict(row) for row in cursor.fetchall()]

    def get_by_id(self, record_id: str, tenant_id: str) -> Optional[Dict[str, Any]]:
        with sqlite3.connect(self.db_path) as conn:
            conn.row_factory = sqlite3.Row
            cursor = conn.execute(
                "SELECT * FROM records WHERE id = ? AND tenant_id = ?",
                (record_id, tenant_id)
            )
            row = cursor.fetchone()
            return dict(row) if row else None
