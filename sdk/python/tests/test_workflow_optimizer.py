"""
Test Suite for Automatic Prompt & Workflow Optimizer (Bounty #37)
Fulfills Acceptance Contract for Hanzo AI Bounty #37 / HIP-1347.
"""

import unittest
from workflow_optimizer import (
    Example,
    CandidateMetrics,
    PromptOptimizer,
    OptimizationRunRecord
)


class TestWorkflowOptimizer(unittest.TestCase):
    def setUp(self):
        self.optimizer = PromptOptimizer(cost_per_1k_tokens_usd=0.002)

        self.train_pool = [
            Example(inputs={"query": "status of cluster 1"}, expected_output="HEALTHY"),
            Example(inputs={"query": "status of node 2"}, expected_output="FAILED"),
            Example(inputs={"query": "status of pod 3"}, expected_output="HEALTHY")
        ]

        self.val_set = [
            Example(inputs={"query": "status of cluster 9"}, expected_output="HEALTHY"),
            Example(inputs={"query": "status of node 8"}, expected_output="FAILED")
        ]

    def test_evaluation_metrics(self):
        # Mock program runner that returns correct output if prompt contains "Verified Demonstrations"
        def mock_runner(instruction: str, inp: dict):
            if "Verified Demonstrations" in instruction:
                # Optimized prompt succeeds on both
                ans = "HEALTHY" if "cluster" in inp["query"] else "FAILED"
                return ans, 150, 15.0
            else:
                # Baseline fails on second item
                return "HEALTHY", 80, 10.0

        base_prompt = "Classify system status into HEALTHY or FAILED."
        metrics = self.optimizer.evaluate_candidate(base_prompt, mock_runner, self.val_set)

        self.assertEqual(metrics.eval_count, 2)
        self.assertEqual(metrics.pass_rate, 0.5)  # 1/2 correct
        self.assertGreater(metrics.cost_usd, 0.0)
        self.assertGreater(metrics.avg_latency_ms, 0.0)

    def test_bootstrap_optimization_run(self):
        def mock_runner(instruction: str, inp: dict):
            if "Verified Demonstrations" in instruction:
                ans = "HEALTHY" if ("cluster" in inp["query"] or "pod" in inp["query"]) else "FAILED"
                return ans, 140, 12.0
            else:
                return "HEALTHY", 70, 8.0

        base_prompt = "You are a monitoring classifier."
        opt_prompt, record = self.optimizer.bootstrap_optimized_prompt(
            base_instruction=base_prompt,
            training_pool=self.train_pool,
            validation_set=self.val_set,
            program_runner=mock_runner,
            max_demonstrations=2
        )

        self.assertIn("Verified Demonstrations", opt_prompt)
        self.assertIsInstance(record, OptimizationRunRecord)
        self.assertEqual(record.baseline_metrics.pass_rate, 0.5)
        self.assertEqual(record.optimized_metrics.pass_rate, 1.0)
        self.assertEqual(record.delta_pass_rate, 0.5)
        self.assertTrue(record.pareto_optimal)
        self.assertIn("opt_run_", record.run_id)


if __name__ == "__main__":
    unittest.main()
