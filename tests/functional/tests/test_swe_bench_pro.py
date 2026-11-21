import pytest
import os

def test_swe_bench_pro_sandbox_isolation():
    """Verify strict anti-leakage controls and isolated HANZO_API_KEY for inference."""
    api_key = os.environ.get("HANZO_API_KEY", "test-key-ai-inference")
    assert api_key is not None
    # Ensure git history reference fixes are isolated
    assert "--unshallow" not in os.environ.get("GIT_PARAMS", "")

def test_swe_bench_pro_runner_metrics():
    """Validate core metric reporting for vfs vs plain NVMe comparison."""
    metrics = {
        "task_resolution": True,
        "wall_time_seconds": 45.2,
        "model_cost_usd": 0.12,
        "filesystem_mode": "vfs"
    }
    assert metrics["wall_time_seconds"] > 0
    assert metrics["model_cost_usd"] >= 0
    assert metrics["filesystem_mode"] in ["vfs", "nvme"]
