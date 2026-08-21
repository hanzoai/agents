package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zap-proto/zip"
)

// This file is the ONE seam between zip's router and the service's existing
// gin handlers. zip owns routing, the middleware chain and the transports;
// the handler bodies stay exactly as they were.
//
// Handlers are fronted rather than rewritten on purpose: the response bodies,
// status codes and streaming semantics of ~24k lines of handler code are this
// service's public API, and a hand port of that volume cannot be proven
// equivalent. The net/http adaptor underneath supports all three response
// modes — buffered, flushed (SSE) and hijacked (WebSocket) — so the streaming
// and WebSocket routes keep working through the same seam.

type (
	// ginParamsKey keys the matched zip route's params, handed to the bridged
	// gin handler through the fasthttp request context.
	ginParamsKey struct{}
	// ginNextKey keys the downstream handler that resumes zip's router once
	// the gin middleware chain declines to abort.
	ginNextKey struct{}
)

// bridgeEngine backs every bridged gin.Context. gin.Context needs an owning
// Engine for allocation and rendering; one shared, route-less engine serves
// all of them. It is never asked to route.
var bridgeEngine = gin.New()

// ginHandler fronts a gin.HandlerFunc as a zip.Handler. The params of the zip
// route that matched are copied onto the gin context, so handlers keep calling
// c.Param("node_id") unchanged.
func ginHandler(h gin.HandlerFunc) zip.Handler {
	serve := zip.AdaptNetHTTP(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := gin.CreateTestContextOnly(w, bridgeEngine)
		c.Request = r
		if p, ok := r.Context().Value(ginParamsKey{}).(gin.Params); ok {
			c.Params = p
		}
		h(c)
		// Mirror gin's engine: flush the status for handlers that set one
		// without writing a body (e.g. c.Status(204)).
		c.Writer.WriteHeaderNow()
	}))

	return func(c *zip.Ctx) error {
		if rt := c.Fiber().Route(); rt != nil && len(rt.Params) > 0 {
			params := make(gin.Params, 0, len(rt.Params))
			for _, name := range rt.Params {
				params = append(params, gin.Param{Key: name, Value: c.Param(name)})
			}
			c.Fiber().RequestCtx().SetUserValue(ginParamsKey{}, params)
		}
		return serve(c)
	}
}

// ginChain fronts the service's gin middleware stack as a single zip
// middleware. The gin middlewares run unmodified, so CORS origin echoing, the
// access-log format and the auth abort bodies are preserved exactly. A chain
// that aborts writes its own response and zip's router is never reached; a
// chain that falls through resumes routing with its headers already applied.
func ginChain(mws ...gin.HandlerFunc) zip.Handler {
	e := gin.New()
	e.Use(mws...)
	// Terminal of the chain. gin reaches this only when no middleware aborted.
	// Setting 200 keeps gin's serveError from emitting its default 404 body:
	// the real response is produced downstream, by zip's router.
	e.NoRoute(func(c *gin.Context) {
		if next, ok := c.Request.Context().Value(ginNextKey{}).(http.Handler); ok {
			next.ServeHTTP(c.Writer, c.Request)
		}
		c.Status(http.StatusOK)
	})

	return zip.AdaptNetHTTPMiddleware(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			e.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), ginNextKey{}, next)))
		})
	})
}
