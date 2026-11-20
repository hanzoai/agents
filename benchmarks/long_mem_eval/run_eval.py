import asyncio
import json
from .harness import LongMemHarness
from .tasks_factory import get_temporal_ordering_task, get_memory_update_task, get_abstention_task

class MockVisor:
    """Mock for local testing before integration with Hanzo Cloud."""
    async def ingest_memory(self, content, timestamp, session_id, policy):
        pass
    
    async def query_agent(self, query, policy):
        class Response:
            text = "Paris" # Mocking a correct answer for temporal
            context_refs = [1]
        return Response()

async def main():
    # Hypothesis: Standard RAG policies without temporal weighting will fail 
    # on 'temporal' and 'update' tasks when context windows are crowded.
    
    visor = MockVisor()
    harness = LongMemHarness(agent_policy_id="context-manager", visor_client=visor)
    
    tasks = [
        get_temporal_ordering_task(),
        get_memory_update_task(),
        get_abstention_task()
    ]
    
    print(f"--- Starting LongMemEval for Hanzo Agents ---")
    for task in tasks:
        result = await harness.run_test_case(task)
        status = "PASS" if result.passed else "FAIL"
        print(f"[{task.category.upper()}] Task {task.id}: {status} ({result.latency_ms:.2f}ms)")

if __name__ == "__main__":
    asyncio.run(main())
