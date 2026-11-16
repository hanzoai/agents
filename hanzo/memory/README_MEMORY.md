# Hanzo Atomic Memory System

This module implements a fact-based memory system inspired by Mem0 and the Kai decision architecture.

## Architecture

1. **Fact Store**: The immutable source of truth. Every piece of information is stored as an atomic `Fact` with a stable identity.
2. **Kai Engine**: A decision-making layer that prevents memory bloat by deduplicating and reconciling new information against the old.
3. **Derived Indexes**: Vector and Lexical indexes are generated from the Fact Store. If the index is corrupted or needs an upgrade (e.g., new embedding model), it can be rebuilt from scratch without data loss.

## Benchmarking

To run benchmarks against LoCoMo or LongMemEval:
1. Use the `MemoryManager.process_event` to ingest the benchmark corpus.
2. Measure `write_amplification` by comparing input tokens vs. stored fact tokens.
3. Measure `retrieval_token_cost` by comparing full transcript retrieval vs. atomic fact retrieval.

## Security
- **Tenant Isolation**: All queries require a `scope` (tenant_id).
- **Access Control**: Filters are applied at the `FactStore` level before any data reaches the retriever.

## License
Apache-2.0
