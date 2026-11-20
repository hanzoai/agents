from .schema import EvalTask, MemoryEvent

def get_temporal_ordering_task() -> EvalTask:
    return EvalTask(
        id="temp_001",
        category="temporal",
        events=[
            MemoryEvent(timestamp=1000, content="Alice visited Paris.", session_id="s1"),
            MemoryEvent(timestamp=2000, content="Alice visited Tokyo.", session_id="s1"),
            MemoryEvent(timestamp=3000, content="Alice visited New York.", session_id="s1"),
        ],
        query="Which city did Alice visit first?",
        expected_answer="Paris"
    )

def get_memory_update_task() -> EvalTask:
    return EvalTask(
        id="upd_001",
        category="update",
        events=[
            MemoryEvent(timestamp=1000, content="The project deadline is Friday.", session_id="s1"),
            MemoryEvent(timestamp=5000, content="The project deadline has been moved to Wednesday.", session_id="s2"),
        ],
        query="When is the project deadline?",
        expected_answer="Wednesday"
    )

def get_abstention_task() -> EvalTask:
    return EvalTask(
        id="abs_001",
        category="abstention",
        events=[
            MemoryEvent(timestamp=1000, content="The company uses Slack for communication.", session_id="s1"),
        ],
        query="What is the company's password for the server?",
        expected_answer="I do not have that information."
    )
