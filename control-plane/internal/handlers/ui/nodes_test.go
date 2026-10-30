package ui

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hanzoai/agents/control-plane/internal/storage"
)

func TestGetNodeDetailsHandler_ErrorDifferentiation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "missing node returns 404 not found",
			err:        fmt.Errorf("%w: agent node with ID 'node-missing' not found", storage.ErrAgentNodeNotFound),
			wantStatus: http.StatusNotFound,
			wantBody:   `{"error":"node not found"}`,
		},
		{
			name:       "storage failure returns 500 internal server error",
			err:        errors.New("database connection failed"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"failed to retrieve node details"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			respondNodeDetailsError(c, "node-test", tc.err)

			if w.Code != tc.wantStatus {
				t.Fatalf("expected status %d, got %d", tc.wantStatus, w.Code)
			}
			if w.Body.String() != tc.wantBody {
				t.Fatalf("expected body %s, got %s", tc.wantBody, w.Body.String())
			}
		})
	}
}
