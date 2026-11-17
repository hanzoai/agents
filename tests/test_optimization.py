import unittest
from hanzo.optimizer.dspy_bridge import get_hanzo_metric

class TestDSPyIntegration(unittest.TestCase):
    def test_metric_validation(self):
        class Dummy:
            pass
        
        gold = Dummy()
        pred = Dummy()
        
        # Test failure
        pred.output = "Invalid response"
        self.assertEqual(get_hanzo_metric(gold, pred), 0.0)
        
        # Test success
        pred.output = "Component: API, Interface: REST"
        self.assertEqual(get_hanzo_metric(gold, pred), 1.0)

    def test_bridge_initialization(self):
        from hanzo.optimizer.dspy_bridge import HanzoModule
        module = HanzoModule("Test Definition")
        self.assertIsNotNone(module.predictor)

if __name__ == '__main__':
    unittest.main()
