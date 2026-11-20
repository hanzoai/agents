import os
import subprocess
from typing import Dict, Any
from .metrics import MetricsCollector

class SWEBenchRunner:
    def __init__(self, sandbox_id: str):
        self.sandbox_id = sandbox_id
        self.collector = MetricsCollector()

    def execute_task(self, task_id: str, config: Dict[str, Any]) -> Dict[str, Any]:
        # Enforce anti-leakage: isolate HANZO_API_KEY
        env = {
            "HANZO_API_KEY": os.getenv("HANZO_API_KEY"),
            "VISOR_SANDBOX_ID": self.sandbox_id,
            "ANTI_LEAKAGE_MODE": "strict"
        }
        
        # Execute via visor CLI
        cmd = ["visor", "run", "--task", task_id, "--vfs", config.get("fs_type", "vfs")]
        
        self.collector.start_timer()
        result = subprocess.run(cmd, env=env, capture_output=True, text=True)
        self.collector.stop_timer()
        
        return {
            "task_id": task_id,
            "success": result.returncode == 0,
            "metrics": self.collector.get_report()
        }
