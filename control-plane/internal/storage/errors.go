package storage

import "errors"

// ErrAgentNodeNotFound is returned when an agent node with the requested ID
// does not exist in storage. Callers can use errors.Is to distinguish a
// missing node from other storage failures.
var ErrAgentNodeNotFound = errors.New("agent node not found")
