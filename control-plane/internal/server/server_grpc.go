//go:build grpc

package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/hanzoai/agents/control-plane/internal/logger"
	"github.com/hanzoai/agents/control-plane/internal/server/middleware"
	"github.com/hanzoai/agents/control-plane/pkg/adminpb"
	"github.com/hanzoai/agents/control-plane/pkg/types"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// startAdminGRPCServer starts the admin gRPC control surface on adminGRPCPort.
// Built only when the `grpc` tag is set; the default build uses the no-op
// stub in server_zap.go.
func (s *HanzoAgentsServer) startAdminGRPCServer() error {
	if s.adminGRPCServer != nil {
		return nil
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.adminGRPCPort))
	if err != nil {
		return err
	}

	s.adminListener = lis
	opts := []grpc.ServerOption{}
	if s.config.API.Auth.APIKey != "" {
		opts = append(opts, grpc.UnaryInterceptor(
			middleware.APIKeyUnaryInterceptor(s.config.API.Auth.APIKey),
		))
	}
	srv := grpc.NewServer(opts...)
	adminpb.RegisterAdminReasonerServiceServer(srv, s)
	s.adminGRPCServer = srv

	go func() {
		if serveErr := srv.Serve(lis); serveErr != nil && !errors.Is(serveErr, grpc.ErrServerStopped) {
			logger.Logger.Error().Err(serveErr).Msg("admin gRPC server stopped unexpectedly")
		}
	}()

	logger.Logger.Info().Int("port", s.adminGRPCPort).Msg("admin gRPC server listening")
	return nil
}

// ListReasoners implements the admin gRPC surface for listing registered reasoners.
func (s *HanzoAgentsServer) ListReasoners(ctx context.Context, _ *adminpb.ListReasonersRequest) (*adminpb.ListReasonersResponse, error) {
	nodes, err := s.storage.ListAgents(ctx, types.AgentFilters{})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list agent nodes: %v", err)
	}

	resp := &adminpb.ListReasonersResponse{}
	for _, node := range nodes {
		if node == nil {
			continue
		}
		for _, reasoner := range node.Reasoners {
			resp.Reasoners = append(resp.Reasoners, &adminpb.Reasoner{
				ReasonerId:    fmt.Sprintf("%s.%s", node.ID, reasoner.ID),
				AgentNodeId:   node.ID,
				Name:          reasoner.ID,
				Description:   fmt.Sprintf("Reasoner %s from node %s", reasoner.ID, node.ID),
				Status:        string(node.HealthStatus),
				NodeVersion:   node.Version,
				LastHeartbeat: node.LastHeartbeat.Format(time.RFC3339),
			})
		}
	}

	return resp, nil
}
