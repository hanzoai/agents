//go:build !grpc

package server

// startAdminGRPCServer is a no-op when the `grpc` build tag is disabled.
// Rebuild with -tags grpc to enable the admin gRPC control surface.
func (s *HanzoAgentsServer) startAdminGRPCServer() error {
	return nil
}
