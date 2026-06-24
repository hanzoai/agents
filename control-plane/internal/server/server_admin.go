package server

import (
	"context"
	"crypto/subtle"
	"fmt"
	"net"
	"time"

	"github.com/hanzoai/agents/control-plane/internal/logger"
	"github.com/hanzoai/agents/control-plane/pkg/types"

	adminwire "github.com/hanzoai/agents/control-plane/pkg/adminwire"
	"github.com/zap-proto/go/rpc"
	"github.com/zap-proto/go/transport"
)

// startAdminServer starts the admin control surface on adminPort over the
// native ZAP transport. The API key, when configured, is carried in each
// request envelope's capability field and checked in constant time before
// dispatch.
func (s *HanzoAgentsServer) startAdminServer() error {
	if s.adminServer != nil {
		return nil
	}

	apiKey := s.config.API.Auth.APIKey

	host := adminBindHost(apiKey)
	if host != "" {
		logger.Logger.Warn().Int("port", s.adminPort).
			Msg("admin server has no API key configured; binding loopback only")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, s.adminPort))
	if err != nil {
		return err
	}

	s.adminServer = transport.Serve(lis, func(env []byte) ([]byte, error) {
		if denied := authorize(env, apiKey); denied != nil {
			return denied, nil
		}
		return adminwire.DispatchAdminReasoner(s, env)
	})

	logger.Logger.Info().Int("port", s.adminPort).Msg("admin server listening")
	return nil
}

// adminBindHost picks the listen host for the admin surface. With no API key
// the surface is unauthenticated, so it binds loopback ("127.0.0.1") and is
// unreachable off-box; with a key gating every call it binds all interfaces
// (""). Reachability and authentication are separate concerns — this decides
// the former from the latter without entangling them.
func adminBindHost(apiKey string) string {
	if apiKey == "" {
		return "127.0.0.1"
	}
	return ""
}

// authorize returns an Unauthorized response envelope when apiKey is set and
// the request's capability does not match it; nil means the call may proceed.
func authorize(env []byte, apiKey string) []byte {
	if apiKey == "" {
		return nil
	}
	call, err := rpc.ParseRequest(env)
	if err != nil {
		return rpc.BuildResponse(rpc.StatusBadRequest, 0, nil)
	}
	if subtle.ConstantTimeCompare(call.Cap, []byte(apiKey)) == 1 {
		return nil
	}
	return rpc.BuildResponse(rpc.StatusUnauthorized, call.PromiseID, nil)
}

// ListReasoners implements adminwire.AdminReasonerHandler: it aggregates the
// reasoners registered across every agent node into one ZAP-encoded list.
func (s *HanzoAgentsServer) ListReasoners() ([]byte, error) {
	ctx := context.Background()
	nodes, err := s.storage.ListAgents(ctx, types.AgentFilters{})
	if err != nil {
		return nil, fmt.Errorf("list agent nodes: %w", err)
	}

	var reasoners [][]byte
	for _, node := range nodes {
		if node == nil {
			continue
		}
		for _, r := range node.Reasoners {
			reasoners = append(reasoners, adminwire.NewReasoner(adminwire.ReasonerInput{
				ReasonerId:    fmt.Sprintf("%s.%s", node.ID, r.ID),
				AgentNodeId:   node.ID,
				Name:          r.ID,
				Description:   fmt.Sprintf("Reasoner %s from node %s", r.ID, node.ID),
				Status:        string(node.HealthStatus),
				NodeVersion:   node.Version,
				LastHeartbeat: node.LastHeartbeat.Format(time.RFC3339),
			}))
		}
	}

	return adminwire.NewListReasonersResponse(adminwire.ListReasonersResponseInput{
		Reasoners: reasoners,
	}), nil
}
