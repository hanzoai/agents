import argparse
import json
import os
from .processor import ZenDatasetGenerator, process_visor_logs

def main():
    parser = argparse.ArgumentParser(description="Hanzo AI Trace Processor")
    parser.add_argument("--input", type=str, required=True, help="Path to visor logs")
    parser.add_argument("--output", type=str, required=True, help="Path to save Zen datasets")
    parser.add_argument("--mode", type=str, default="sft", choices=["sft", "dpo"])
    
    args = parser.parse_args()
    
    generator = ZenDatasetGenerator(pii_masking=True)
    trajectories = process_visor_logs(args.input)
    
    output_data = []
    for traj in trajectories:
        if args.mode == "sft":
            output_data.extend(generator.to_sft_format(traj))
            
    os.makedirs(os.path.dirname(args.output), exist_ok=True)
    with open(args.output, 'w') as f:
        for entry in output_data:
            f.write(json.dumps(entry) + "\n")
            
    print(f"Processed {len(trajectories)} trajectories. Output saved to {args.output}")

if __name__ == "__main__":
    main()
