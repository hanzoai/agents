package workflow

import (
	"context"
	"time"
)

type EventType string

const (
	EventTypeActivityStarted   EventType = "ActivityStarted"
	EventTypeActivityCompleted EventType = "ActivityCompleted"
	EventTypeActivityFailed    EventType = "ActivityFailed"
	EventTypeWorkflowStarted   EventType = "WorkflowStarted"
	EventTypeWorkflowCompleted EventType = "WorkflowCompleted"
)

type HistoryEvent struct {
	ID             string      `json:"id"`
	Type           EventType   `json:"type"`
	Timestamp      time.Time   `json:"timestamp"`
	ActivityName   string      `json:"activity_name,omitempty"`
	IdempotencyKey string      `json:"idempotency_key,omitempty"`
	Input          interface{} `json:"input,omitempty"`
	Result         interface{} `json:"result,omitempty"`
	Error          string      `json:"error,omitempty"`
}

type WorkflowState struct {
	WorkflowID string         `json:"workflow_id"`
	History    []HistoryEvent `json:"history"`
	Status     string         `json:"status"`
}

type ActivityOptions struct {
	IdempotencyKey string
	Timeout        time.Duration
	RetryPolicy    *RetryPolicy
}

type RetryPolicy struct {
	InitialInterval    time.Duration
	BackoffCoefficient float64
	MaximumAttempts    int
}
