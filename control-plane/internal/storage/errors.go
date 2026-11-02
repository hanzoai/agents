package storage

import "errors"

// ErrAgentNodeNotFound is returned (wrapped with %w) when an agent node
// lookup fails because the requested node does not exist. Callers should use
// errors.Is(err, ErrAgentNodeNotFound) to distinguish a missing node (-> 404)
// from other storage failures (-> 500).
var ErrAgentNodeNotFound = errors.New("agent node not found")
