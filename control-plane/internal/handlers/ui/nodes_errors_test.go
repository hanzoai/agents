package ui

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hanzoai/agents/control-plane/internal/core/domain"
	"github.com/hanzoai/agents/control-plane/internal/core/interfaces"
	"github.com/hanzoai/agents/control-plane/internal/services"
	"github.com/hanzoai/agents/control-plane/internal/storage"
	"github.com/hanzoai/agents/control-plane/pkg/types"

	"github.com/gin-gonic/gin"
)

// ---- stubs (fully self-contained; no dependency on other *_test.go files) ----

// nodeErrStore implements storage.StorageProvider. The embedded nil interface
// supplies all unused methods; only GetAgent is called on the node-details path.
type nodeErrStore struct {
	storage.StorageProvider
	err error
}

func (s *nodeErrStore) GetAgent(ctx context.Context, id string) (*types.AgentNode, error) {
	return nil, s.err
}

// stubAgentClient satisfies interfaces.AgentClient; methods are never called.
type stubAgentClient struct{}

func (stubAgentClient) GetMCPHealth(ctx context.Context, nodeID string) (*interfaces.MCPHealthResponse, error) {
	return nil, nil
}
func (stubAgentClient) RestartMCPServer(ctx context.Context, nodeID, alias string) error { return nil }
func (stubAgentClient) GetMCPTools(ctx context.Context, nodeID, alias string) (*interfaces.MCPToolsResponse, error) {
	return nil, nil
}
func (stubAgentClient) ShutdownAgent(ctx context.Context, nodeID string, graceful bool, timeoutSeconds int) (*interfaces.AgentShutdownResponse, error) {
	return nil, nil
}
func (stubAgentClient) GetAgentStatus(ctx context.Context, nodeID string) (*interfaces.AgentStatusResponse, error) {
	return nil, nil
}

// stubAgentService satisfies interfaces.AgentService; methods are never called.
type stubAgentService struct{}

func (stubAgentService) RunAgent(name string, options domain.RunOptions) (*domain.RunningAgent, error) {
	return nil, nil
}
func (stubAgentService) StopAgent(name string) error { return nil }
func (stubAgentService) GetAgentStatus(name string) (*domain.AgentStatus, error) {
	return nil, nil
}
func (stubAgentService) ListRunningAgents() ([]domain.RunningAgent, error) { return nil, nil }

// ---- test harness ----

func newNodeDetailsRouter(t *testing.T, getAgentErr error) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := &nodeErrStore{err: getAgentErr}
	client := stubAgentClient{}
	statusManager := services.NewStatusManager(store, services.StatusManagerConfig{}, nil, client)
	uiService := services.NewUIService(store, client, stubAgentService{}, statusManager)

	handler := NewNodesHandler(uiService)
	r := gin.New()
	r.GET("/nodes/:nodeId", handler.GetNodeDetailsHandler)
	return r
}

func doNodeDetails(r *gin.Engine) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/nodes/some-id", nil)
	r.ServeHTTP(w, req)
	return w
}

// ---- assertions ----

func TestGetNodeDetailsHandler_NotFound(t *testing.T) {
	r := newNodeDetailsRouter(t, storage.ErrAgentNodeNotFound)
	w := doNodeDetails(r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestGetNodeDetailsHandler_NotFoundWrapped(t *testing.T) {
	wrapped := &wrappedErr{prefix: "agent node with ID 'x': ", err: storage.ErrAgentNodeNotFound}
	r := newNodeDetailsRouter(t, wrapped)
	w := doNodeDetails(r)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404 for wrapped sentinel, got %d (body: %s)", w.Code, w.Body.String())
	}
}

func TestGetNodeDetailsHandler_ServerError(t *testing.T) {
	r := newNodeDetailsRouter(t, errors.New("database is down"))
	w := doNodeDetails(r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("want 500, got %d (body: %s)", w.Code, w.Body.String())
	}
}

type wrappedErr struct {
	prefix string
	err    error
}

func (w *wrappedErr) Error() string { return w.prefix + w.err.Error() }
func (w *wrappedErr) Unwrap() error { return w.err }
