import unittest
from tools.trace_processor.processor import ZenDatasetGenerator
from tools.trace_processor.models import Trajectory, TraceStep

class TestTraceProcessor(unittest.TestCase):
    def setUp(self):
        self.generator = ZenDatasetGenerator(pii_masking=True)
        self.mock_step = TraceStep(
            step_id="1",
            thought="I should list files",
            action="ls -la",
            observation="file1.py",
            reward=1.0
        )
        self.mock_trajectory = Trajectory(
            trajectory_id="test-123",
            agent_id="coder-agent",
            steps=[self.mock_step],
            success=True,
            total_reward=1.0
        )

    def test_pii_masking(self):
        raw_text = "My key is sk-1234567890abcdef1234567890abcdef"
        masked = self.generator.mask_content(raw_text)
        self.assertIn("[MASKED]", masked)
        self.assertNotIn("sk-1234567890", masked)

    def test_sft_conversion(self):
        dataset = self.generator.to_sft_format(self.mock_trajectory)
        self.assertEqual(len(dataset), 1)
        self.assertEqual(dataset[0]["output"], "ls -la")

if __name__ == "__main__":
    unittest.main()
