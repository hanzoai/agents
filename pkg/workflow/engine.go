package workflow

import (
	"context"
	"time"
)

type Engine struct {
	Storage HistoryStorage
}

func NewEngine(storage HistoryStorage) *Engine {
	return &Engine{Storage: storage}
}

func (e *Engine) Run(ctx context.Context, workflowID string, logic func(wctx *WorkflowContext) error) error {
	state, err := e.Storage.LoadState(workflowID)
	if err != nil {
		return err
	}

	var history []HistoryEvent
	if state != nil {
		history = state.History
	} else {
		history = []HistoryEvent{
			{
				Type:      EventTypeWorkflowStarted,
				Timestamp: time.Now(),
			},
		}
	}

	wctx := &WorkflowContext{
		Context:    ctx,
		workflowID: workflowID,
		history:    history,
		storage:    e.Storage,
		isReplaying: len(history) > 1,
	}

	err = logic(wctx)
	
	if err == nil {
		finalEvent := HistoryEvent{
			Type:      EventTypeWorkflowCompleted,
			Timestamp: time.Now(),
		}
		wctx.newEvents = append(wctx.newEvents, finalEvent)
		return wctx.persist()
	}

	return err
}
