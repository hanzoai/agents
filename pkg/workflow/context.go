package workflow

import (
	"context"
	"fmt"
)

type WorkflowContext struct {
	context.Context
	workflowID     string
	history        []HistoryEvent
	replayIndex    int
	isReplaying    bool
	newEvents      []HistoryEvent
	storage        HistoryStorage
}

func (w *WorkflowContext) ExecuteActivity(name string, opts ActivityOptions, fn func() (interface{}, error)) (interface{}, error) {
	// Check if we have this in history (Replay Phase)
	for i := 0; i < len(w.history); i++ {
		event := w.history[i]
		if event.Type == EventTypeActivityCompleted && event.IdempotencyKey == opts.IdempotencyKey {
			// Skip execution, return recorded result
			return event.Result, nil
		}
	}

	// Execution Phase
	w.isReplaying = false
	
	startEvent := HistoryEvent{
		ID:             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		Type:           EventTypeActivityStarted,
		Timestamp:      time.Now(),
		ActivityName:   name,
		IdempotencyKey: opts.IdempotencyKey,
	}
	w.newEvents = append(w.newEvents, startEvent)

	result, err := fn()
	
	if err != nil {
		failEvent := HistoryEvent{
			Type:           EventTypeActivityFailed,
			Timestamp:      time.Now(),
			ActivityName:   name,
			IdempotencyKey: opts.IdempotencyKey,
			Error:          err.Error(),
		}
		w.newEvents = append(w.newEvents, failEvent)
		_ = w.persist()
		return nil, err
	}

	completeEvent := HistoryEvent{
		Type:           EventTypeActivityCompleted,
		Timestamp:      time.Now(),
		ActivityName:   name,
		IdempotencyKey: opts.IdempotencyKey,
		Result:         result,
	}
	w.newEvents = append(w.newEvents, completeEvent)
	
	err = w.persist()
	return result, err
}

func (w *WorkflowContext) persist() error {
	w.history = append(w.history, w.newEvents...)
	w.newEvents = []HistoryEvent{}
	state := WorkflowState{
		WorkflowID: w.workflowID,
		History:    w.history,
		Status:     "RUNNING",
	}
	return w.storage.SaveState(w.workflowID, state)
}
