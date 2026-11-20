import time
from typing import List, Dict
from .schema import EvalTask, EvalResult, MemoryEvent

class LongMemHarness:
    """
    Objective evaluation harness for existing Hanzo memory builders.
    Interfaces with the canonical stack: vfs -> hanzoai/s3 -> visor.
    """
    
    def __init__(self, agent_policy_id: str, visor_client):
        self.policy_id = agent_policy_id
        self.visor = visor_client # Abstracted Hanzo Visor interface

    async def run_test_case(self, task: EvalTask) -> EvalResult:
        start_time = time.time()
        
        # 1. Inject Memory Events into the VFS/S3 via Visor
        for event in task.events:
            await self.visor.ingest_memory(
                content=event.content,
                timestamp=event.timestamp,
                session_id=event.session_id,
                policy=self.policy_id
            )
        
        # 2. Query the Agent (Zen/Kai routing)
        # Note: We expect the agent to use the memory policy defined in #62-#66
        response = await self.visor.query_agent(
            query=task.query,
            policy=self.policy_id
        )
        
        latency = (time.time() - start_time) * 1000
        
        # 3. Evaluate Correctness
        passed = self._evaluate_logic(task, response.text)
        
        return EvalResult(
            task_id=task.id,
            passed=passed,
            model_output=response.text,
            latency_ms=latency,
            retrieved_context_ids=response.context_refs
        )

    def _evaluate_logic(self, task: EvalTask, output: str) -> bool:
        output_lower = output.lower()
        expected_lower = task.expected_answer.lower()
        
        if task.category == "abstention":
            # Check if the agent correctly refused to hallucinate
            return any(phrase in output_lower for phrase in ["i don't know", "not mentioned", "no information"])
        
        if task.category == "update":
            # Ensure the LATEST information is used, not the stale one
            return expected_lower in output_lower
            
        return expected_lower in output_lower
