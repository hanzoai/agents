import pytest
from tau2_bench.dual_control.engine import DualControlEngine, ActionSource, StateConflictError

def test_dual_control_concurrency():
    engine = DualControlEngine()
    entity_id = "order_123"
    
    # Human locks state
    engine.mutate(entity_id, "LOCKED_BY_HUMAN", ActionSource.HUMAN)
    
    # Agent attempts mutation
    with pytest.raises(StateConflictError):
        engine.mutate(entity_id, "PROCESSED", ActionSource.AGENT)
        
    # Verify final state integrity
    final_state = engine.store[entity_id]
    assert final_state.status == "LOCKED_BY_HUMAN"
    assert final_state.version == 1
