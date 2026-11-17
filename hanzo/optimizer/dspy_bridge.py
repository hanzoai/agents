import dspy
from typing import Dict, Any

class HanzoAgentSignature(dspy.Signature):
    """
    Dynamic signature for Hanzo Agents.
    Maps to Zen (input/output) and Kai (structured decisions).
    """
    context = dspy.InputField(desc="The system context and agent definition")
    task = dspy.InputField(desc="The specific task to be performed")
    output = dspy.OutputField(desc="The structured response or code output")

class HanzoModule(dspy.Module):
    def __init__(self, agent_definition: str):
        super().__init__()
        self.agent_definition = agent_definition
        self.predictor = dspy.Predict(HanzoAgentSignature)

    def forward(self, task: str):
        result = self.predictor(context=self.agent_definition, task=task)
        return dspy.Prediction(output=result.output)

def get_hanzo_metric(gold: Any, prediction: Any, trace=None):
    """
    Metric to validate Kai's structured decisions.
    Checks for valid JSON and specific key presence.
    """
    predicted_output = prediction.output
    # Logic to interface with Kai's validation
    if not predicted_output:
        return False
    
    # Example: Check if output contains required architectural components
    score = 1.0 if all(key in predicted_output.lower() for key in ["component", "interface"]) else 0.0
    return score
