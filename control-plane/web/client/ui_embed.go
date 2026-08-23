//go:build embedded

// UI embedding and route registration for HanzoAgents (embedded build).

package client

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"strings"

	"github.com/zap-proto/zip"
)

//go:embed dist/* dist/**
var UIFiles embed.FS

// isStaticAsset reports whether path looks like a web asset rather than an SPA
// route. This prevents reasoner IDs with dots (like
// "deepresearchagent.meta_research_methodology_reasoner") from being treated as
// static assets.
func isStaticAsset(path string) bool {
	for _, ext := range []string{
		".js", ".css", ".html", ".ico", ".png", ".jpg", ".jpeg", ".gif", ".svg",
		".woff", ".woff2", ".ttf", ".eot", ".map", ".json", ".xml", ".txt",
	} {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}

// RegisterUIRoutes registers the UI routes with the zip app.
func RegisterUIRoutes(app *zip.App) {
	fmt.Println("Registering embedded UI routes...")

	// Create a sub-filesystem that strips the "dist" prefix
	uiFS, err := fs.Sub(UIFiles, "dist")
	if err != nil {
		panic("Failed to create UI filesystem: " + err.Error())
	}

	// One handler for the whole SPA. zip.Static reads the file from the "*"
	// capture, serves index.html for the root and for directories, and falls
	// back to it for a path that names no file — which is what a client-side
	// route is. That was three hand-written branches, an extension allowlist and
	// a net/http FileServer behind an adapter; the extension list in particular
	// decided by SUFFIX what only the filesystem can answer, so a real asset with
	// an unlisted extension was served the app shell instead of itself.
	//
	// It is also stricter than what it replaces: traversal fails closed, and
	// Content-Type, Content-Length, Last-Modified, HEAD and If-Modified-Since
	// are all handled rather than left to the caller.
	ui := zip.Static(uiFS, zip.WithIndex("index.html"), zip.WithFallback("index.html"))
	app.Get("/ui/*", ui)
	app.Head("/ui/*", ui)

	app.Get("/", func(c *zip.Ctx) error {
		return c.Redirect(http.StatusMovedPermanently, "/ui/")
	})

	// Fallback - serve index.html for /ui/* routes that don't match static
	// files, 404 JSON for everything else. Least-specific route, so every
	// registered route still wins. "+" rather than "*" so it does not shadow
	// the root redirect registered above.
	app.All("/+", func(c *zip.Ctx) error {
		return c.JSON(http.StatusNotFound, map[string]any{"error": "endpoint not found"})
	})
}

// IsUIEmbedded checks if UI files are embedded in the binary.
func IsUIEmbedded() bool {
	// Try to read a file that should exist in the embedded UI
	_, err := UIFiles.ReadFile("dist/index.html")
	return err == nil
}
