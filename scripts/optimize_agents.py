import os
import dspy
import argparse
import json
from dspy.teleprompt import BootstrapFewShot
from hanzo.optimizer.dspy_bridge import HanzoModule, get_hanzo_metric

def load_agent_definition(agent_name: str) -> str:
    path = f".hanzo/agents/definitions/{agent_name}.md"
    with open(path, 'r') as f:
        return f.read()

def run_optimization(agent_name: str, trainset_path: str):
    # Setup LLM (Enso routing equivalent)
    lm = dspy.OpenAI(model='gpt-4', max_tokens=2000)
    dspy.settings.configure(lm=lm)

    # Load Data
    with open(trainset_path, 'r') as f:
        data = json.load(f)
        trainset = [dspy.Example(task=d['task'], output=d['output']).with_inputs('task') for d in data]

    # Initialize Module
    agent_def = load_agent_definition(agent_name)
    hanzo_program = HanzoModule(agent_def)

    # Setup Optimizer
    optimizer = BootstrapFewShot(metric=get_hanzo_metric)
    
    print(f"Optimizing agent: {agent_name}...")
    optimized_program = optimizer.compile(hanzo_program, trainset=trainset)

    # Save Results
    output_path = f"research/optimization_{agent_name}.json"
    optimized_program.save(output_path)
    print(f"Optimization complete. Saved to {output_path}")

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--agent", type=str, required=True)
    parser.add_argument("--dataset", type=str, required=True)
    args = parser.parse_args()
    
    run_optimization(args.agent, args.dataset)
