package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/hanzoai/orm/relational"
)

// openEngine binds the relational engine to driverName/dsn and hands its
// connection pool back to the caller as the store's only *sql.DB, so the ORM
// and the hand-written SQL in this package share one pool and one transaction
// scope. Times are read and written in UTC, matching what the raw statements
// store.
func openEngine(driverName, dsn string) (*relational.Engine, error) {
	engine, err := relational.NewEngine(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s engine: %w", driverName, err)
	}
	engine.TZLocation = time.UTC
	engine.DatabaseTZ = time.UTC
	return engine, nil
}

// session returns an auto-closing engine session bound to ctx.
func (ls *LocalStorage) session(ctx context.Context) (*relational.Session, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}
	if ls.engine == nil {
		return nil, fmt.Errorf("relational engine is not initialized")
	}
	return ls.engine.Context(ctx), nil
}
