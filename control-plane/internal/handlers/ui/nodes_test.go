package ui

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hanzoai/agents/control-plane/internal/services"
	"github.com/hanzoai/agents/control-plane/internal/storage"
	"github.com/hanzoai/agents/control-plane/pkg/types"
)

// stubNodeStorage implements storage.StorageProvider for testing the full
// Handler -> UIService.GetNodeDetailsWithPackageInfo -> Storage.GetAgent path.
type stubNodeStorage struct {
	storage.StorageProvider
	getAgentFn           func(ctx context.Context, id string) (*types.AgentNode, error)
	queryAgentPackagesFn func(ctx context.Context, filters types.PackageFilters) ([]*types.AgentPackage, error)
}

func (s *stubNodeStorage) GetAgent(ctx context.Context, id string) (*types.AgentNode, error) {
	if s.getAgentFn != nil {
		return s.getAgentFn(ctx, id)
	}
	return nil, fmt.Errorf("%w: agent node with ID '%s' not found", storage.ErrAgentNodeNotFound, id)
}

func (s *stubNodeStorage) QueryAgentPackages(ctx context.Context, filters types.PackageFilters) ([]*types.AgentPackage, error) {
	if s.queryAgentPackagesFn != nil {
		return s.queryAgentPackagesFn(ctx, filters)
	}
	return nil, nil
}

func TestGetNodeDetailsHandler_EndToEndThroughServiceAndStorage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		nodeID     string
		getAgentFn func(ctx context.Context, id string) (*types.AgentNode, error)
		wantStatus int
		wantBody   string
	}{
		{
			name:   "missing node in storage propagates ErrAgentNodeNotFound through UIService and returns 404",
			nodeID: "node-missing",
			getAgentFn: func(ctx context.Context, id string) (*types.AgentNode, error) {
				return nil, fmt.Errorf("%w: agent node with ID '%s' not found", storage.ErrAgentNodeNotFound, id)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"node not found"}`,
		},
		{
			name:   "database failure in storage propagates through UIService and returns 500",
			nodeID: "node-db-failure",
			getAgentFn: func(ctx context.Context, id string) (*types.AgentNode, error) {
				return nil, fmt.Errorf("failed to get agent node with ID '%s': %w", id, errors.New("sqlite disk I/O error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"failed to retrieve node details"}`,
		},
		{
			name:   "existing node in storage returns 200 OK through UIService and Handler",
			nodeID: "node-healthy",
			getAgentFn: func(ctx context.Context, id string) (*types.AgentNode, error) {
				return &types.AgentNode{
					ID:      id,
					TeamID:  "team-core",
					Version: "1.0.0",
				}, nil
			},
			wantStatus: http.StatusOK,
			wantBody:   "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			store := &stubNodeStorage{getAgentFn: tc.getAgentFn}
			uiService := services.NewUIService(store, nil, nil, nil)
			handler := NewNodesHandler(uiService)

			router := gin.New()
			router.GET("/v1/ui/nodes/:nodeId/details", handler.GetNodeDetailsHandler)

			req := httptest.NewRequest(http.MethodGet, "/v1/ui/nodes/"+tc.nodeID+"/details", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d (body: %s)", tc.wantStatus, w.Code, w.Body.String())
			}
			if tc.wantBody != "" && w.Body.String() != tc.wantBody {
				t.Fatalf("expected body %s, got %s", tc.wantBody, w.Body.String())
			}
		})
	}
}
