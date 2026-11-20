package workflow

import (
	"context"
	"errors"
	"os"
	"testing"
)

func TestDurableWorkflowReplay(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "workflow-test")
	defer os.RemoveAll(tmpDir)

	storage := &VFSStorage{BasePath: tmpDir}
	engine := NewEngine(storage)
	workflowID := "test-workflow-123"

	activityCount := 0
	
	// Define workflow logic
	logic := func(wctx *WorkflowContext) error {
		// Activity 1
		_, err := wctx.ExecuteActivity("Step1", ActivityOptions{IdempotencyKey: "step-1"}, func() (interface{}, error) {
			activityCount++
			return "result-1", nil
		})
		if err != nil {
			return err
		}

		// Activity 2 - This one will fail the first time
		_, err = wctx.ExecuteActivity("Step2", ActivityOptions{IdempotencyKey: "step-2"}, func() (interface{}, error) {
			if activityCount < 2 {
				return nil, errors.New("simulated failure")
			}
			activityCount++
			return "result-2", nil
		})
		return err
	}

	// First Run: Should fail at Step 2
	err := engine.Run(context.Background(), workflowID, logic)
	if err == nil || err.Error() != "simulated failure" {
		t.Errorf("Expected simulated failure, got %v", err)
	}

	if activityCount != 1 {
		t.Errorf("Expected 1 activity execution, got %d", activityCount)
	}

	// Second Run: Should replay Step 1 (skip) and execute Step 2
	err = engine.Run(context.Background(), workflowID, logic)
	if err != nil {
		t.Errorf("Expected success on retry, got %v", err)
	}

	// Total activityCount should be 2 (Step 1 was skipped during replay)
	if activityCount != 2 {
		t.Errorf("Replay failed: Activity 1 was re-executed. Total count: %d", activityCount)
	}
}
