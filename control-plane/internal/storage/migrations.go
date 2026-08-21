package storage

import (
	"context"
	"fmt"

	"github.com/hanzoai/orm/relational"
)

func (ls *LocalStorage) autoMigrateSchema(ctx context.Context) error {
	sess, err := ls.session(ctx)
	if err != nil {
		return fmt.Errorf("failed to prepare session for migrations: %w", err)
	}

	if ls.mode == "local" {
		if _, err := ls.db.Exec("PRAGMA foreign_keys = OFF"); err != nil {
			return fmt.Errorf("failed to disable foreign keys: %w", err)
		}
		defer func() {
			if _, err := ls.db.Exec("PRAGMA foreign_keys = ON"); err != nil {
				fmt.Printf("failed to re-enable foreign keys: %v\n", err)
			}
		}()
	}

	models := []interface{}{
		&ExecutionRecordModel{},
		&AgentExecutionModel{},
		&AgentNodeModel{},
		&AgentConfigurationModel{},
		&AgentPackageModel{},
		&WorkflowExecutionModel{},
		&WorkflowExecutionEventModel{},
		&WorkflowRunEventModel{},
		&WorkflowRunModel{},
		&WorkflowStepModel{},
		&WorkflowModel{},
		&SessionModel{},
		&DIDRegistryModel{},
		&AgentDIDModel{},
		&ComponentDIDModel{},
		&ExecutionVCModel{},
		&WorkflowVCModel{},
		&SchemaMigrationModel{},
		&ExecutionWebhookEventModel{},
		&ExecutionWebhookModel{},
		&ObservabilityWebhookModel{},
		&ObservabilityDeadLetterQueueModel{},
	}

	// Additive only. The indexes this package creates from raw DDL — and the
	// ones migrations/*.sql add — are described by no struct, so letting the
	// engine drop what it does not recognise would delete them on every boot.
	if _, err := sess.SyncWithOptions(relational.SyncOptions{IgnoreDropIndices: true}, models...); err != nil {
		return fmt.Errorf("failed to sync schema: %w", err)
	}

	return nil
}
