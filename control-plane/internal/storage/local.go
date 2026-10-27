package storage

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hanzoai/agents/control-plane/internal/events"
	"github.com/hanzoai/agents/control-plane/pkg/types"

	"github.com/boltdb/bolt"
	"github.com/hanzoai/orm/relational"
	"github.com/hanzoai/sqlite"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // Import pgx driver for PostgreSQL
)

// Custom error types for data integrity issues
type DuplicateDIDError struct {
	DID  string
	Type string // "registry", "agent", or "component"
}

func (e *DuplicateDIDError) Error() string {
	return fmt.Sprintf("duplicate %s DID detected: %s already exists", e.Type, e.DID)
}

// ForeignKeyConstraintError represents a foreign key constraint violation
type ForeignKeyConstraintError struct {
	Table           string
	Column          string
	ReferencedTable string
	ReferencedValue string
	Operation       string
}

func (e *ForeignKeyConstraintError) Error() string {
	return fmt.Sprintf("foreign key constraint violation in %s.%s: referenced %s '%s' does not exist (operation: %s)",
		e.Table, e.Column, e.ReferencedTable, e.ReferencedValue, e.Operation)
}

// ValidationError represents a pre-storage validation failure
type ValidationError struct {
	Field   string
	Value   string
	Reason  string
	Context string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s='%s': %s (context: %s)",
		e.Field, e.Value, e.Reason, e.Context)
}

// getWorkflowExecutionByID is a helper function that retrieves a workflow execution using DBTX interface
func (ls *LocalStorage) getWorkflowExecutionByID(ctx context.Context, q DBTX, executionID string) (*types.WorkflowExecution, error) {
	query := `
		SELECT workflow_id, execution_id, hanzo_agents_request_id, run_id, session_id, actor_id,
		       agent_node_id, parent_workflow_id, parent_execution_id, root_workflow_id, workflow_depth,
		       reasoner_id, input_data, output_data, input_size, output_size,
		       status, started_at, completed_at, duration_ms,
		       state_version, last_event_sequence, active_children, pending_children,
		       pending_terminal_status, status_reason, lease_owner, lease_expires_at,
		       error_message, retry_count, workflow_name, workflow_tags, notes, created_at, updated_at
		FROM workflow_executions WHERE execution_id = ?`

	row := q.QueryRowContext(ctx, query, executionID)
	execution := &types.WorkflowExecution{}

	var workflowTagsJSON, notesJSON []byte
	var inputData, outputData sql.NullString
	var runID sql.NullString
	var pendingTerminal sql.NullString
	var statusReason sql.NullString
	var leaseOwner sql.NullString
	var leaseExpires sql.NullTime
	err := row.Scan(
		&execution.WorkflowID, &execution.ExecutionID, &execution.HanzoAgentsRequestID,
		&runID, &execution.SessionID, &execution.ActorID, &execution.AgentNodeID,
		&execution.ParentWorkflowID, &execution.ParentExecutionID, &execution.RootWorkflowID, &execution.WorkflowDepth,
		&execution.ReasonerID, &inputData, &outputData,
		&execution.InputSize, &execution.OutputSize, &execution.Status,
		&execution.StartedAt, &execution.CompletedAt, &execution.DurationMS,
		&execution.StateVersion, &execution.LastEventSequence, &execution.ActiveChildren, &execution.PendingChildren,
		&pendingTerminal, &statusReason,
		&leaseOwner, &leaseExpires,
		&execution.ErrorMessage, &execution.RetryCount, &execution.WorkflowName,
		&workflowTagsJSON, &notesJSON, &execution.CreatedAt, &execution.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// "Not found" is a valid case for an upsert operation, so we return nil without an error.
			// The caller is responsible for handling the nil execution record.
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get workflow execution: %w", err)
	}

	// Handle nullable JSON fields
	if runID.Valid {
		execution.RunID = &runID.String
	}
	if inputData.Valid {
		execution.InputData = safeJSONRawMessage(inputData.String, "{}", fmt.Sprintf("execution %s input_data", execution.ExecutionID))
	} else {
		execution.InputData = json.RawMessage("{}")
	}

	if outputData.Valid {
		execution.OutputData = safeJSONRawMessage(outputData.String, "{}", fmt.Sprintf("execution %s output_data", execution.ExecutionID))
	} else {
		execution.OutputData = json.RawMessage("{}")
	}
	if pendingTerminal.Valid {
		execution.PendingTerminalStatus = &pendingTerminal.String
	}
	if statusReason.Valid {
		execution.StatusReason = &statusReason.String
	}
	if leaseOwner.Valid {
		execution.LeaseOwner = &leaseOwner.String
	}
	if leaseExpires.Valid {
		t := leaseExpires.Time
		execution.LeaseExpiresAt = &t
	}

	// Unmarshal workflow tags
	if len(workflowTagsJSON) > 0 {
		if err := json.Unmarshal(workflowTagsJSON, &execution.WorkflowTags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal workflow tags: %w", err)
		}
	}

	// Unmarshal notes
	if len(notesJSON) > 0 {
		if err := json.Unmarshal(notesJSON, &execution.Notes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal notes: %w", err)
		}
	} else {
		execution.Notes = []types.ExecutionNote{}
	}

	return execution, nil
}

func (ls *LocalStorage) StoreWorkflowRun(ctx context.Context, run *types.WorkflowRun) error {
	if run == nil {
		return fmt.Errorf("workflow run cannot be nil")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	db := ls.requireSQLDB()
	createdAt := run.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := run.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}
	metadata := "{}"
	if len(run.Metadata) > 0 {
		metadata = string(run.Metadata)
	}

	query := `
		INSERT INTO workflow_runs (
			run_id, root_workflow_id, root_execution_id, status, total_steps,
			completed_steps, failed_steps, state_version, last_event_sequence,
			metadata, created_at, updated_at, completed_at
		) VALUES (?, ?, ?, ?, ?, ?,
