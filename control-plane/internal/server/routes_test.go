package server

import (
	"fmt"
	"net/http"
	"sort"
	"testing"

	"github.com/hanzoai/agents/control-plane/internal/config"

	"github.com/stretchr/testify/require"
)

// wantRoutes is the control-plane's route table, captured from the gin router
// this service was migrated off. It is the contract: the zip router must expose
// exactly these method+path pairs, no more and no fewer. The only spelling
// change from gin is the wildcard leaf ("/ui/*filepath" -> "/ui/*"), which
// matches the same paths.
var wantRoutes = []string{
	"DELETE /v1/memory/vector/:key",
	"DELETE /v1/memory/vector/namespace",
	"DELETE /v1/nodes/:node_id/monitoring",
	"DELETE /v1/settings/observability-webhook",
	"DELETE /v1/settings/observability-webhook/dlq",
	"DELETE /v1/ui/agents/:agentId/env/:key",
	"DELETE /v1/ui/workflows/:workflowId/cleanup",
	"GET /",
	"GET /health",
	"GET /metrics",
	"GET /ui/*",
	"GET /v1/discovery/capabilities",
	"GET /v1/executions/:execution_id",
	"GET /v1/executions/:execution_id/notes",
	"GET /v1/health",
	"GET /v1/memory/events/history",
	"GET /v1/memory/events/sse",
	"GET /v1/memory/events/ws",
	"GET /v1/memory/list",
	"GET /v1/memory/vector/:key",
	"GET /v1/nodes",
	"GET /v1/nodes/:node_id",
	"GET /v1/nodes/:node_id/status",
	"GET /v1/settings/observability-webhook",
	"GET /v1/settings/observability-webhook/dlq",
	"GET /v1/settings/observability-webhook/status",
	"GET /v1/ui/agents/:agentId/config",
	"GET /v1/ui/agents/:agentId/config/schema",
	"GET /v1/ui/agents/:agentId/details",
	"GET /v1/ui/agents/:agentId/env",
	"GET /v1/ui/agents/:agentId/executions",
	"GET /v1/ui/agents/:agentId/executions/:executionId",
	"GET /v1/ui/agents/:agentId/status",
	"GET /v1/ui/agents/packages",
	"GET /v1/ui/agents/packages/:packageId/details",
	"GET /v1/ui/agents/running",
	"GET /v1/ui/dashboard/enhanced",
	"GET /v1/ui/dashboard/summary",
	"GET /v1/ui/did/:did/resolution-bundle",
	"GET /v1/ui/did/:did/resolution-bundle/download",
	"GET /v1/ui/did/export/vcs",
	"GET /v1/ui/did/status",
	"GET /v1/ui/executions/:execution_id/details",
	"GET /v1/ui/executions/:execution_id/notes",
	"GET /v1/ui/executions/:execution_id/vc",
	"GET /v1/ui/executions/:execution_id/vc-status",
	"GET /v1/ui/executions/enhanced",
	"GET /v1/ui/executions/events",
	"GET /v1/ui/executions/recent",
	"GET /v1/ui/executions/stats",
	"GET /v1/ui/executions/summary",
	"GET /v1/ui/executions/timeline",
	"GET /v1/ui/identity/agents",
	"GET /v1/ui/identity/agents/:agent_id/details",
	"GET /v1/ui/identity/credentials/search",
	"GET /v1/ui/identity/dids/search",
	"GET /v1/ui/identity/dids/stats",
	"GET /v1/ui/mcp/status",
	"GET /v1/ui/nodes/:nodeId/details",
	"GET /v1/ui/nodes/:nodeId/did",
	"GET /v1/ui/nodes/:nodeId/mcp/events",
	"GET /v1/ui/nodes/:nodeId/mcp/health",
	"GET /v1/ui/nodes/:nodeId/mcp/metrics",
	"GET /v1/ui/nodes/:nodeId/mcp/servers/:alias/tools",
	"GET /v1/ui/nodes/:nodeId/status",
	"GET /v1/ui/nodes/:nodeId/vc-status",
	"GET /v1/ui/nodes/events",
	"GET /v1/ui/nodes/summary",
	"GET /v1/ui/reasoners/:reasonerId/details",
	"GET /v1/ui/reasoners/:reasonerId/executions",
	"GET /v1/ui/reasoners/:reasonerId/metrics",
	"GET /v1/ui/reasoners/:reasonerId/templates",
	"GET /v1/ui/reasoners/all",
	"GET /v1/ui/reasoners/events",
	"GET /v1/ui/vc/:vcId/download",
	"GET /v1/ui/workflows/:workflowId/dag",
	"GET /v1/ui/workflows/:workflowId/notes/events",
	"GET /v1/ui/workflows/:workflowId/vc-chain",
	"GET /v2/ui/workflow-runs",
	"GET /v2/ui/workflow-runs/:run_id",
	"HEAD /ui/*",
	"PATCH /v1/nodes/:node_id/status",
	"PATCH /v1/ui/agents/:agentId/env",
	"POST /v1/actions/claim",
	"POST /v1/execute/:target",
	"POST /v1/execute/async/:target",
	"POST /v1/executions/:execution_id/status",
	"POST /v1/executions/batch-status",
	"POST /v1/executions/note",
	"POST /v1/memory/delete",
	"POST /v1/memory/get",
	"POST /v1/memory/set",
	"POST /v1/memory/vector",
	"POST /v1/memory/vector/delete",
	"POST /v1/memory/vector/search",
	"POST /v1/memory/vector/set",
	"POST /v1/nodes",
	"POST /v1/nodes/:node_id/actions/ack",
	"POST /v1/nodes/:node_id/heartbeat",
	"POST /v1/nodes/:node_id/lifecycle/status",
	"POST /v1/nodes/:node_id/shutdown",
	"POST /v1/nodes/:node_id/start",
	"POST /v1/nodes/:node_id/status/refresh",
	"POST /v1/nodes/:node_id/stop",
	"POST /v1/nodes/register",
	"POST /v1/nodes/register-serverless",
	"POST /v1/nodes/status/bulk",
	"POST /v1/nodes/status/refresh",
	"POST /v1/reasoners/:reasoner_id",
	"POST /v1/settings/observability-webhook",
	"POST /v1/settings/observability-webhook/presets/console",
	"POST /v1/settings/observability-webhook/redrive",
	"POST /v1/skills/:skill_id",
	"POST /v1/ui/agents/:agentId/config",
	"POST /v1/ui/agents/:agentId/reconcile",
	"POST /v1/ui/agents/:agentId/start",
	"POST /v1/ui/agents/:agentId/stop",
	"POST /v1/ui/executions/:execution_id/verify-vc",
	"POST /v1/ui/executions/:execution_id/webhook/retry",
	"POST /v1/ui/executions/note",
	"POST /v1/ui/nodes/:nodeId/mcp/servers/:alias/restart",
	"POST /v1/ui/nodes/:nodeId/status/refresh",
	"POST /v1/ui/nodes/status/bulk",
	"POST /v1/ui/nodes/status/refresh",
	"POST /v1/ui/reasoners/:reasonerId/templates",
	"POST /v1/ui/vc/verify",
	"POST /v1/ui/workflows/:workflowId/verify-vc",
	"POST /v1/ui/workflows/vc-status",
	"POST /v1/workflow/executions/events",
	"PUT /v1/ui/agents/:agentId/env",
}

// gotRoutes returns the app's real route table: middleware entries and the
// NoRoute catch-all excluded, deduplicated, sorted.
func gotRoutes(srv *HanzoAgentsServer) (routes, catchAll []string) {
	seen := map[string]bool{}
	for _, r := range srv.App.Fiber().GetRoutes(true) {
		key := fmt.Sprintf("%s %s", r.Method, r.Path)
		if seen[key] {
			continue
		}
		seen[key] = true
		if r.Path == "/+" {
			catchAll = append(catchAll, key)
			continue
		}
		routes = append(routes, key)
	}
	sort.Strings(routes)
	sort.Strings(catchAll)
	return routes, catchAll
}

// TestRouteTableParity pins the migrated route table to the pre-migration one.
func TestRouteTableParity(t *testing.T) {
	t.Parallel()

	srv := newTestServer(config.UIConfig{Enabled: true, Mode: "embedded"}, config.APIConfig{})
	srv.setupRoutes()

	got, catchAll := gotRoutes(srv)

	require.Equal(t, wantRoutes, got, "route table drifted from the pre-migration contract")
	require.Len(t, got, 130, "control-plane serves 130 routes")
	require.NotEmpty(t, catchAll, "the NoRoute catch-all must be registered")
}

// TestNoRouteFallback pins gin's NoRoute semantics onto zip's catch-all: the
// root route still wins "/", and unmatched paths get the same 404 body.
func TestNoRouteFallback(t *testing.T) {
	t.Parallel()

	srv := newTestServer(config.UIConfig{Enabled: true, Mode: "embedded"}, config.APIConfig{})
	srv.setupRoutes()

	t.Run("root is not shadowed by the catch-all", func(t *testing.T) {
		code, hdr, _ := doReq(t, srv, http.MethodGet, "/", nil)
		require.Equal(t, http.StatusMovedPermanently, code)
		require.Equal(t, "/ui/", hdr.Get("Location"))
	})

	t.Run("unknown non-ui path returns the gin NoRoute body", func(t *testing.T) {
		code, _, body := doReq(t, srv, http.MethodGet, "/v1/does-not-exist", nil)
		require.Equal(t, http.StatusNotFound, code)
		require.JSONEq(t, `{"error":"endpoint not found"}`, body)
	})

	t.Run("registered route still wins over the catch-all", func(t *testing.T) {
		code, _, _ := doReq(t, srv, http.MethodGet, "/health", nil)
		require.Equal(t, http.StatusOK, code)
	})
}
