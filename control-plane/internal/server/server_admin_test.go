package server

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hanzoai/agents/control-plane/internal/storage"
	"github.com/hanzoai/agents/control-plane/pkg/types"

	adminwire "github.com/hanzoai/agents/control-plane/pkg/adminwire"
	"github.com/stretchr/testify/require"
	"github.com/zap-proto/go/rpc"
)

func TestListReasonersAggregatesNodes(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tempDir := t.TempDir()
	cfg := storage.StorageConfig{
		Mode: "local",
		Local: storage.LocalStorageConfig{
			DatabasePath: filepath.Join(tempDir, "hanzo-agents.db"),
			KVStorePath:  filepath.Join(tempDir, "hanzo-agents.bolt"),
		},
	}

	localStore := storage.NewLocalStorage(storage.LocalStorageConfig{})
	if err := localStore.Initialize(ctx, cfg); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "fts5") {
			t.Skip("sqlite3 compiled without FTS5; skipping reasoner aggregation test")
		}
		require.NoError(t, err)
	}
	t.Cleanup(func() { _ = localStore.Close(ctx) })

	srv := &HanzoAgentsServer{storage: localStore}

	schema := json.RawMessage("{}")
	node := &types.AgentNode{
		ID:            "node-1",
		HealthStatus:  types.HealthStatusActive,
		Version:       "1.0.0",
		LastHeartbeat: time.Now().UTC(),
		Reasoners: []types.ReasonerDefinition{
			{ID: "reason", InputSchema: schema, OutputSchema: schema},
			{ID: "another", InputSchema: schema, OutputSchema: schema},
		},
	}
	require.NoError(t, localStore.RegisterAgent(ctx, node))

	body, err := srv.ListReasoners()
	require.NoError(t, err)

	resp, err := adminwire.WrapListReasonersResponse(body)
	require.NoError(t, err)
	list := resp.Reasoners()
	require.Equal(t, 2, list.Len())

	first, err := adminwire.WrapReasoner(list.BytesAt(0))
	require.NoError(t, err)
	require.Equal(t, "node-1.reason", first.ReasonerId())
	require.Equal(t, "node-1", first.AgentNodeId())
	require.NotEmpty(t, first.LastHeartbeat())
}

func TestAuthorize(t *testing.T) {
	t.Parallel()

	listReq := func(cap string) []byte {
		return rpc.BuildRequest(rpc.Call{
			Method:    adminwire.AdminReasonerListReasonersOrdinal,
			PromiseID: 7,
			Cap:       []byte(cap),
		})
	}

	t.Run("no key configured allows any cap", func(t *testing.T) {
		require.Nil(t, authorize(listReq(""), ""))
		require.Nil(t, authorize(listReq("anything"), ""))
	})

	t.Run("matching cap proceeds", func(t *testing.T) {
		require.Nil(t, authorize(listReq("secret-key"), "secret-key"))
	})

	t.Run("missing cap is unauthorized", func(t *testing.T) {
		denied := authorize(listReq(""), "secret-key")
		require.NotNil(t, denied)
		resp, err := rpc.ParseResponse(denied)
		require.NoError(t, err)
		require.Equal(t, rpc.StatusUnauthorized, resp.Status)
		require.Equal(t, uint32(7), resp.PromiseID)
	})

	t.Run("wrong cap is unauthorized", func(t *testing.T) {
		denied := authorize(listReq("wrong-key"), "secret-key")
		require.NotNil(t, denied)
		resp, err := rpc.ParseResponse(denied)
		require.NoError(t, err)
		require.Equal(t, rpc.StatusUnauthorized, resp.Status)
	})

	t.Run("malformed envelope is bad request", func(t *testing.T) {
		denied := authorize([]byte("not a zap envelope"), "secret-key")
		require.NotNil(t, denied)
		resp, err := rpc.ParseResponse(denied)
		require.NoError(t, err)
		require.Equal(t, rpc.StatusBadRequest, resp.Status)
	})
}

func TestAdminBindHost(t *testing.T) {
	t.Parallel()

	// Unauthenticated (no key) must bind loopback so the surface is not
	// reachable off-box; an authenticated surface binds all interfaces.
	require.Equal(t, "127.0.0.1", adminBindHost(""), "no key must bind loopback only")
	require.Equal(t, "", adminBindHost("secret-key"), "with a key, bind all interfaces")
}
