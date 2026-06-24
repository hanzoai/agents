// Code for SCIENTIST verification of the gRPC->ZAP cut. Proves the admin
// reasoner surface round-trips over the REAL native ZAP transport (a real
// socket), end to end: client builds a ZAP request envelope, transport frames
// it across the socket, DispatchAdminReasoner routes ordinal 1 to the handler,
// the handler encodes a canonical ZAP ListReasonersResponse, the transport
// frames it back, and the client decodes the typed Reasoner views with every
// field intact. Independent of the server's sqlite/FTS5 storage wiring on
// purpose: the dispatch + wire are what is under test.
package admin

import (
	"net"
	"testing"

	"github.com/zap-proto/go/rpc"
	"github.com/zap-proto/go/transport"
)

// fixedReasoners is a hand-built handler returning a known reasoner list. It
// satisfies AdminReasonerHandler (the same contract HanzoAgentsServer.ListReasoners
// implements) without touching storage.
type fixedReasoners struct{}

func (fixedReasoners) ListReasoners() ([]byte, error) {
	r0 := NewReasoner(ReasonerInput{
		ReasonerId:    "node-1.alpha",
		AgentNodeId:   "node-1",
		Name:          "alpha",
		Description:   "Reasoner alpha from node node-1",
		Status:        "active",
		NodeVersion:   "1.2.3",
		LastHeartbeat: "2026-06-24T00:00:00Z",
	})
	r1 := NewReasoner(ReasonerInput{
		ReasonerId:    "node-2.beta",
		AgentNodeId:   "node-2",
		Name:          "beta",
		Description:   "Reasoner beta from node node-2",
		Status:        "draining",
		NodeVersion:   "4.5.6",
		LastHeartbeat: "2026-06-24T01:02:03Z",
	})
	return NewListReasonersResponse(ListReasonersResponseInput{
		Reasoners: [][]byte{r0, r1},
	}), nil
}

// serveRoundTrip stands up a real transport.Serve on the given network/addr,
// dials it with the generated AdminReasonerClient, calls ListReasoners, and
// asserts the two-element list round-trips with all id/name/status/version/
// heartbeat fields intact.
func serveRoundTrip(t *testing.T, network, addr string) {
	t.Helper()

	ln, err := net.Listen(network, addr)
	if err != nil {
		t.Fatalf("listen %s %s: %v", network, addr, err)
	}
	// transport.Serve is the EXACT registration server_admin.go uses.
	srv := transport.Serve(ln, func(env []byte) ([]byte, error) {
		return DispatchAdminReasoner(fixedReasoners{}, env)
	})
	defer srv.Close()

	// Dial over the real socket; *transport.Conn satisfies AdminReasonerChannel.
	conn, err := transport.Dial(network, srv.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := NewAdminReasonerClient(conn, nil)
	_, body, err := client.ListReasoners()
	if err != nil {
		t.Fatalf("ListReasoners RPC: %v", err)
	}

	resp, err := WrapListReasonersResponse(body)
	if err != nil {
		t.Fatalf("decode response: %v", err)
	}
	list := resp.Reasoners()
	if got := list.Len(); got != 2 {
		t.Fatalf("reasoner count = %d, want 2", got)
	}

	type want struct{ id, node, name, status, version, hb string }
	wants := []want{
		{"node-1.alpha", "node-1", "alpha", "active", "1.2.3", "2026-06-24T00:00:00Z"},
		{"node-2.beta", "node-2", "beta", "draining", "4.5.6", "2026-06-24T01:02:03Z"},
	}
	for i, w := range wants {
		r, err := WrapReasoner(list.BytesAt(i))
		if err != nil {
			t.Fatalf("WrapReasoner[%d]: %v", i, err)
		}
		if r.ReasonerId() != w.id {
			t.Errorf("[%d] ReasonerId = %q, want %q", i, r.ReasonerId(), w.id)
		}
		if r.AgentNodeId() != w.node {
			t.Errorf("[%d] AgentNodeId = %q, want %q", i, r.AgentNodeId(), w.node)
		}
		if r.Name() != w.name {
			t.Errorf("[%d] Name = %q, want %q", i, r.Name(), w.name)
		}
		if r.Status() != w.status {
			t.Errorf("[%d] Status = %q, want %q", i, r.Status(), w.status)
		}
		if r.NodeVersion() != w.version {
			t.Errorf("[%d] NodeVersion = %q, want %q", i, r.NodeVersion(), w.version)
		}
		if r.LastHeartbeat() != w.hb {
			t.Errorf("[%d] LastHeartbeat = %q, want %q", i, r.LastHeartbeat(), w.hb)
		}
	}
}

func TestScientistRoundTripTCP(t *testing.T) {
	serveRoundTrip(t, "tcp", "127.0.0.1:0")
}

func TestScientistRoundTripUnix(t *testing.T) {
	sock := t.TempDir() + "/admin.sock"
	serveRoundTrip(t, "unix", sock)
}

// TestScientistUnknownOrdinal confirms attack-vector 3: an unknown method
// ordinal over the same transport yields StatusNotFound, not a panic or a
// silent hang. Built by hand because the generated client only knows ordinal 1.
func TestScientistUnknownOrdinal(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := transport.Serve(ln, func(env []byte) ([]byte, error) {
		return DispatchAdminReasoner(fixedReasoners{}, env)
	})
	defer srv.Close()

	conn, err := transport.Dial("tcp", srv.Addr().String())
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	req := rpc.BuildRequest(rpc.Call{Method: 999, PromiseID: 42})
	resp, err := conn.Call(req)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if resp.Status != rpc.StatusNotFound {
		t.Fatalf("unknown ordinal status = %d, want %d (StatusNotFound)", resp.Status, rpc.StatusNotFound)
	}
	if resp.PromiseID != 42 {
		t.Fatalf("PromiseID = %d, want 42 (echoed)", resp.PromiseID)
	}
	if len(resp.Body) != 0 {
		t.Fatalf("NotFound body = %d bytes, want empty", len(resp.Body))
	}
}
