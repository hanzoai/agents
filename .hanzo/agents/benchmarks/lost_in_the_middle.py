import json
import random

class ContextRobustnessBenchmark:
    """
    Benchmark suite to evaluate evidence-ordering robustness.
    Based on 'Lost in the Middle' (Liu et al., 2024).
    """
    def __init__(self, agent_context_manager):
        self.manager = agent_context_manager

    def generate_test_case(self, num_docs, target_pos):
        """
        Generates a synthetic context where the ground truth is placed at target_pos.
        """
        docs = [f"Irrelevant document content {i}" for i in range(num_docs)]
        ground_truth = "The secret key is 42."
        docs.insert(target_pos, ground_truth)
        return docs

    def run_evaluation(self, model_interface):
        results = {"beginning": 0, "middle": 0, "end": 0}
        positions = [0, len(self.generate_test_case(10, 0)) // 2, 9]
        
        for pos in positions:
            context = self.generate_test_case(10, pos)
            response = model_interface.query(context, "What is the secret key?")
            if "42" in response:
                results[self._pos_name(pos)] += 1
        return results

    def _pos_name(self, pos):
        if pos == 0: return "beginning"
        return "end" if pos > 5 else "middle"
