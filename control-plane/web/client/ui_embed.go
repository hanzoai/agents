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

// serveIndex writes the embedded SPA entry point.
func serveIndex(c *zip.Ctx) error {
	indexHTML, err := UIFiles.ReadFile("dist/index.html")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"error": "Failed to load UI index",
		})
	}
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	return c.Bytes(http.StatusOK, indexHTML)
}

// RegisterUIRoutes registers the UI routes with the zip app.
func RegisterUIRoutes(app *zip.App) {
	fmt.Println("Registering embedded UI routes...")

	// Create a sub-filesystem that strips the "dist" prefix
	uiFS, err := fs.Sub(UIFiles, "dist")
	if err != nil {
		panic("Failed to create UI filesystem: " + err.Error())
	}

	fileServer := zip.AdaptNetHTTP(http.StripPrefix("/ui", http.FileServer(http.FS(uiFS))))

	ui := func(c *zip.Ctx) error {
		path := strings.TrimPrefix(c.Path(), "/ui")

		// If accessing root UI path or a directory, serve index.html
		if path == "/" || path == "" || strings.HasSuffix(path, "/") {
			return serveIndex(c)
		}

		if isStaticAsset(strings.ToLower(path)) {
			return fileServer(c)
		}

		// For all other paths (SPA routes), serve index.html
		return serveIndex(c)
	}

	app.Get("/ui/*", ui)
	app.Head("/ui/*", ui)

	// Root redirect to embedded UI
	app.Get("/", func(c *zip.Ctx) error {
		return c.Redirect(http.StatusMovedPermanently, "/ui/")
	})

	// Fallback - serve index.html for /ui/* routes that don't match static
	// files, 404 JSON for everything else. Least-specific route, so every
	// registered route still wins. "+" rather than "*" so it does not shadow
	// the root redirect registered above.
	app.All("/+", func(c *zip.Ctx) error {
		if strings.HasPrefix(c.Path(), "/ui/") {
			return serveIndex(c)
		}
		return c.JSON(http.StatusNotFound, map[string]any{"error": "endpoint not found"})
	})
}

// IsUIEmbedded checks if UI files are embedded in the binary.
func IsUIEmbedded() bool {
	// Try to read a file that should exist in the embedded UI
	_, err := UIFiles.ReadFile("dist/index.html")
	return err == nil
}
