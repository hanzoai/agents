package storage

import "errors"

// ErrAgentNodeNotFound reports that no agent node exists with the requested
// ID. Callers that need to tell a missing record apart from a broken backend
// check it with errors.Is; plain backend failures are wrapped with %w too but
// never match this sentinel.
var ErrAgentNodeNotFound = errors.New("not found")
