//go:build !embedded

package client

import "github.com/zap-proto/zip"

// RegisterUIRoutes is a no-op when the UI is not embedded.
// The server will serve UI assets from the filesystem when configured.
func RegisterUIRoutes(_ *zip.App) {}

// IsUIEmbedded reports whether UI assets are embedded in the binary.
func IsUIEmbedded() bool { return false }
