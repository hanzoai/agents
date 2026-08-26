package server

import (
	"net/http"
	"testing"

	"github.com/hanzoai/agents/control-plane/internal/config"

	"github.com/stretchr/testify/require"
)

// TestMetricsExposition pins what GET /metrics answers, so the route keeps
// answering it however it is written. The version rides in the media type and a
// scraper reads it to choose a parser, so the type is pinned whole rather than
// by prefix. The deadline a scraper asks for is part of the answer: it arrives
// under either spelling, a value that is not a number means no deadline, and a
// deadline already spent is refused rather than answered with a partial body.
func TestMetricsExposition(t *testing.T) {
	t.Parallel()

	srv := newTestServer(config.UIConfig{Enabled: true, Mode: "embedded"}, config.APIConfig{})
	srv.setupRoutes()

	served := []struct {
		name   string
		header map[string]string
	}{
		{"no deadline asked for", nil},
		{"a deadline", map[string]string{"X-Scrape-Timeout-Seconds": "5"}},
		{"a deadline, spelled the older way", map[string]string{"X-Prometheus-Scrape-Timeout-Seconds": "5"}},
		{"a deadline that is not a number", map[string]string{"X-Scrape-Timeout-Seconds": "soon"}},
	}
	for _, s := range served {
		t.Run(s.name, func(t *testing.T) {
			code, hdr, body := doReq(t, srv, http.MethodGet, "/metrics", s.header)

			require.Equal(t, http.StatusOK, code)
			require.Equal(t, "text/plain; version=0.0.4; charset=utf-8", hdr.Get("Content-Type"))
			require.Contains(t, body, "# HELP hanzo_agents_gateway_queue_depth ")
			require.Contains(t, body, "# TYPE hanzo_agents_gateway_queue_depth gauge")
		})
	}

	t.Run("a deadline already spent", func(t *testing.T) {
		code, hdr, body := doReq(t, srv, http.MethodGet, "/metrics", map[string]string{"X-Scrape-Timeout-Seconds": "1e-9"})

		require.Equal(t, http.StatusInternalServerError, code)
		require.Equal(t, "text/plain; charset=utf-8", hdr.Get("Content-Type"))
		require.Equal(t, "nosniff", hdr.Get("X-Content-Type-Options"))
		require.Equal(t, "metrics gather error\n", body)
	})
}
