package server

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/zap-proto/fiber/v3"
	"github.com/zap-proto/zip"
)

// TestGinBridge pins the seam between zip's router and the service's gin
// handlers: path params, status codes, response bodies, the gin middleware
// chain (CORS origin echo + auth abort) and SSE streaming must all survive it
// unchanged. These are the properties the whole migration rests on.
func TestGinBridge(t *testing.T) {
	t.Parallel()

	app := zip.New(zip.Config{DisableStartupMessage: true})

	app.Use(ginChain(
		cors.New(cors.Config{
			AllowOrigins: []string{"http://localhost:3000", "http://localhost:5173"},
			AllowMethods: []string{"GET", "POST"},
			AllowHeaders: []string{"Origin", "X-API-Key"},
		}),
		func(c *gin.Context) {
			if c.Request.URL.Path == "/secret" && c.GetHeader("X-API-Key") != "k" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
			c.Next()
		},
	))

	app.Get("/nodes/:node_id/status", ginHandler(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"id": c.Param("node_id")})
	}))
	app.Get("/secret", ginHandler(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}))
	app.Get("/teapot", ginHandler(func(c *gin.Context) {
		c.JSON(http.StatusTeapot, gin.H{"err": "nope"})
	}))
	app.Get("/stream", ginHandler(func(c *gin.Context) {
		c.Header("Content-Type", "text/event-stream")
		f, _ := c.Writer.(http.Flusher)
		for i := 0; i < 3; i++ {
			fmt.Fprintf(c.Writer, "data: e%d\n\n", i)
			f.Flush()
		}
	}))

	do := func(method, path string, hdr map[string]string) (*http.Response, string) {
		req := httptest.NewRequest(method, path, nil)
		for k, v := range hdr {
			req.Header.Set(k, v)
		}
		resp, err := app.Fiber().Test(req, fiber.TestConfig{Timeout: 10 * time.Second, FailOnTimeout: true})
		if err != nil {
			t.Fatalf("%s %s: %v", method, path, err)
		}
		b, _ := io.ReadAll(resp.Body)
		return resp, string(b)
	}

	t.Run("param+body", func(t *testing.T) {
		resp, body := do("GET", "/nodes/abc123/status", nil)
		if resp.StatusCode != 200 || body != `{"id":"abc123"}` {
			t.Fatalf("got %d %q", resp.StatusCode, body)
		}
	})

	t.Run("status preserved", func(t *testing.T) {
		resp, body := do("GET", "/teapot", nil)
		if resp.StatusCode != http.StatusTeapot || body != `{"err":"nope"}` {
			t.Fatalf("got %d %q", resp.StatusCode, body)
		}
	})

	t.Run("cors echoes origin", func(t *testing.T) {
		resp, _ := do("GET", "/teapot", map[string]string{"Origin": "http://localhost:5173"})
		if got := resp.Header.Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Fatalf("ACAO = %q", got)
		}
	})

	t.Run("auth abort", func(t *testing.T) {
		resp, body := do("GET", "/secret", nil)
		if resp.StatusCode != 401 || body != `{"error":"unauthorized"}` {
			t.Fatalf("got %d %q", resp.StatusCode, body)
		}
		resp, body = do("GET", "/secret", map[string]string{"X-API-Key": "k"})
		if resp.StatusCode != 200 || body != `{"ok":true}` {
			t.Fatalf("authed: got %d %q", resp.StatusCode, body)
		}
	})

	t.Run("sse streams", func(t *testing.T) {
		resp, body := do("GET", "/stream", nil)
		if resp.StatusCode != 200 || body != "data: e0\n\ndata: e1\n\ndata: e2\n\n" {
			t.Fatalf("got %d %q", resp.StatusCode, body)
		}
		if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
			t.Fatalf("content-type = %q", ct)
		}
	})
}
