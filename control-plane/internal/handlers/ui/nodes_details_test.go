package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hanzoai/agents/control-plane/internal/services"
	"github.com/hanzoai/agents/control-plane/internal/storage"
	"github.com/hanzoai/agents/control-plane/pkg/types"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockStorageForNodeDetails struct {
	storage.StorageProvider
	agentNode *types.AgentNode
	err       error
}

func (m *mockStorageForNodeDetails) GetAgent(ctx context.Context, id string) (*types.AgentNode, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.agentNode, nil
}

func (m *mockStorageForNodeDetails) QueryAgentPackages(ctx context.Context, filters types.PackageFilters) ([]*types.AgentPackage, error) {
	return nil, nil
}

func TestGetNodeDetailsHandler_NotFoundVsStorageError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockAgentClient := unreachableAgent()
	mockAgentService := &MockAgentServiceForUI{}

	t.Run("missing node returns 404 not found", func(t *testing.T) {
		store := &mockStorageForNodeDetails{
			err: fmt.Errorf("agent node with ID 'missing-node': %w", storage.ErrAgentNodeNotFound),
		}
		statusManager := services.NewStatusManager(store, services.StatusManagerConfig{}, nil, mockAgentClient)
		uiService := services.NewUIService(store, mockAgentClient, mockAgentService, statusManager)
		handler := NewNodesHandler(uiService)

		router := gin.New()
		router.GET("/v1/ui/nodes/:nodeId", handler.GetNodeDetailsHandler)

		req := httptest.NewRequest(http.MethodGet, "/v1/ui/nodes/missing-node", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.Contains(t, strings.ToLower(resp.Body.String()), "not found")
	})

	t.Run("storage failure returns 500 internal server error", func(t *testing.T) {
		store := &mockStorageForNodeDetails{
			err: errors.New("database connection refused"),
		}
		statusManager := services.NewStatusManager(store, services.StatusManagerConfig{}, nil, mockAgentClient)
		uiService := services.NewUIService(store, mockAgentClient, mockAgentService, statusManager)
		handler := NewNodesHandler(uiService)

		router := gin.New()
		router.GET("/v1/ui/nodes/:nodeId", handler.GetNodeDetailsHandler)

		req := httptest.NewRequest(http.MethodGet, "/v1/ui/nodes/node-1", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		assert.Contains(t, resp.Body.String(), "failed to retrieve node details")
	})

	t.Run("existing node returns 200 ok", func(t *testing.T) {
		store := &mockStorageForNodeDetails{
			agentNode: &types.AgentNode{
				ID:      "node-123",
				BaseURL: "http://localhost:8080",
			},
		}
		statusManager := services.NewStatusManager(store, services.StatusManagerConfig{}, nil, mockAgentClient)
		uiService := services.NewUIService(store, mockAgentClient, mockAgentService, statusManager)
		handler := NewNodesHandler(uiService)

		router := gin.New()
		router.GET("/v1/ui/nodes/:nodeId", handler.GetNodeDetailsHandler)

		req := httptest.NewRequest(http.MethodGet, "/v1/ui/nodes/node-123", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "node-123")
	})
}
