package ui

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hanzoai/agents/control-plane/internal/services"
	"github.com/hanzoai/agents/control-plane/internal/storage"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newNodesTestRouter builds GetNodeDetailsHandler over the given storage.
func newNodesTestRouter(t *testing.T, store storage.StorageProvider) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mockAgentClient := unreachableAgent()
	mockAgentService := &MockAgentServiceForUI{}
	statusManager := services.NewStatusManager(store, services.StatusManagerConfig{}, nil, mockAgentClient)
	uiService := services.NewUIService(store, mockAgentClient, mockAgentService, statusManager)

	handler := NewNodesHandler(uiService)
	router := gin.New()
	router.GET("/v1/ui/nodes/:nodeId", handler.GetNodeDetailsHandler)
	return router
}

// newTestStorage returns an initialized local storage backed by a temp dir.
func newTestStorage(t *testing.T) *storage.LocalStorage {
	t.Helper()
	ctx := context.Background()
	tempDir := t.TempDir()
	cfg := storage.StorageConfig{
		Mode: "local",
		Local: storage.LocalStorageConfig{
			DatabasePath: tempDir + "/test.db",
			KVStorePath:  tempDir + "/test.bolt",
		},
	}

	realStorage := storage.NewLocalStorage(storage.LocalStorageConfig{})
	err := realStorage.Initialize(ctx, cfg)
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "fts5") {
		t.Skip("sqlite3 compiled without FTS5")
	}
	require.NoError(t, err)
	t.Cleanup(func() { _ = realStorage.Close(ctx) })
	return realStorage
}

func doNodeRequest(router *gin.Engine, id string) *httptest.ResponseRecorder {
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, httptest.NewRequest(http.MethodGet, "/v1/ui/nodes/"+id, nil))
	return resp
}

// A node id that does not exist must stay a 404 with a "not found" message.
func TestGetNodeDetailsHandler_MissingNodeReturns404(t *testing.T) {
	store := newTestStorage(t)
	router := newNodesTestRouter(t, store)

	resp := doNodeRequest(router, "no-such-node")
	assert.Equal(t, http.StatusNotFound, resp.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	assert.Contains(t, body["error"], "not found")
}

// A storage failure is a server fault: 500, not 404, and no internals leak.
func TestGetNodeDetailsHandler_StorageFailureReturns500(t *testing.T) {
	store := newTestStorage(t)
	router := newNodesTestRouter(t, store)

	// Break the backend so every read fails with a plain, non-sentinel error
	// ("sql: database is closed") instead of a missing row.
	require.NoError(t, store.Close(context.Background()))

	resp := doNodeRequest(router, "any-node")
	assert.Equal(t, http.StatusInternalServerError, resp.Code)

	var body map[string]string
	require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &body))
	assert.Contains(t, body["error"], "failed to retrieve node details")
	assert.NotContains(t, body["error"], "database is closed") // no internals
}

// The storage layer must actually report a missing row with the sentinel so
// the handler can classify it with errors.Is.
func TestGetAgent_MissingNodeWrapsSentinel(t *testing.T) {
	store := newTestStorage(t)

	_, err := store.GetAgent(context.Background(), "no-such-node")
	require.Error(t, err)
	assert.ErrorIs(t, err, storage.ErrAgentNodeNotFound)
}
