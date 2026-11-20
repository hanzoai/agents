import time
import psutil

class MetricsCollector:
    def __init__(self):
        self.start_time = 0
        self.end_time = 0

    def start_timer(self):
        self.start_time = time.perf_counter()

    def stop_timer(self):
        self.end_time = time.perf_counter()

    def get_report(self) -> dict:
        return {
            "wall_time": self.end_time - self.start_time,
            "io_read_bytes": psutil.disk_io_counters().read_bytes,
            "io_write_bytes": psutil.disk_io_counters().write_bytes
        }
