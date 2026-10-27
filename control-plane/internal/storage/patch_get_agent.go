package storage

// PATCH INSTRUCTIONS for local.go GetAgent method (around line 4196):
//
// This file exists only as documentation. The actual change must be made in local.go.
//
// In the GetAgent function, change the "not found" error from:
//
//   return nil, fmt.Errorf("agent node with ID '%s' not found", id)
//
// to:
//
//   return nil, fmt.Errorf("agent node with ID '%s': %w", id, ErrAgentNodeNotFound)
//
// This wraps the sentinel ErrAgentNodeNotFound so callers can use errors.Is().
// The other error path (database/scan failures) should remain unchanged,
// returning a generic error that does NOT wrap ErrAgentNodeNotFound.
