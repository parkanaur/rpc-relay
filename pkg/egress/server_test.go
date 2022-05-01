package egress

import (
	"context"
	"encoding/json"
	gnatsd "github.com/nats-io/gnatsd/server"
	natstest "github.com/nats-io/nats-server/test"
	"github.com/nats-io/nats.go"
	"github.com/parkanaur/rpc-relay/pkg/jrpcserver"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func NewJRPCServer(t *testing.T, config *relayutil.Config) *http.Server {
	srv, err := jrpcserver.NewServer(config)
	if err != nil {
		t.Fatal(err)
	}
	httpSrv := &http.Server{Addr: config.JRPCServer.GetHostWithPort(), Handler: srv}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}()
	return httpSrv
}

type RelayFixture struct {
	NATSTestServer *gnatsd.Server
	JRPCServer     *http.Server
	EgressServer   *Server
}

type RPCCalcSumResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  int    `json:"result"`
	ID      int    `json:"id"`
}

func NewRelayFixture(t *testing.T, config *relayutil.Config) (*RelayFixture, error) {
	natsSrv := natstest.RunServer(nil)
	jrpcSrv := NewJRPCServer(t, config)
	egrSrv, err := NewServer(config)
	if err != nil {
		return nil, err
	}
	return &RelayFixture{natsSrv, jrpcSrv, egrSrv}, nil
}

func (fixture *RelayFixture) Shutdown() error {
	err := fixture.EgressServer.Shutdown()
	if err != nil {
		return err
	}

	err = fixture.JRPCServer.Shutdown(context.Background())
	if err != nil {
		return err
	}

	fixture.NATSTestServer.Shutdown()
	return nil
}

func TestNewServer(t *testing.T) {
	natsSrv := natstest.RunServer(nil)
	defer natsSrv.Shutdown()
	cf := relayutil.NewTestConfig()
	server, err := NewServer(cf)
	assert.NoError(t, err)

	assert.NotNil(t, server.NATSConnection)
	assert.Equal(t, server.NATSConnection.ConnectedUrl(), cf.NATS.ServerURL)
	assert.Equal(t, server.NATSConnection.Status(), nats.CONNECTED)

	assert.NotNil(t, server.RPCClient)

	assert.Equal(t, server.config, cf)

	_ = server.Shutdown()
}

func TestServer_Shutdown(t *testing.T) {
	natsSrv := natstest.RunServer(nil)
	defer natsSrv.Shutdown()
	cf := relayutil.NewTestConfig()
	server, _ := NewServer(cf)

	err := server.Shutdown()
	assert.NoError(t, err)

	assert.Equal(t, server.NATSConnection.Status(), nats.CLOSED)
}

func TestServer_handleRPCRequest(t *testing.T) {
	cf := relayutil.NewTestConfig()
	fixture, err := NewRelayFixture(t, cf)
	if err != nil {
		t.Fatal(err)
	}
	defer fixture.Shutdown()

	rq, err := fixture.EgressServer.NATSConnection.Request(
		"rpc.calculateSum.calculateSum",
		[]byte(`{"jsonrpc": "2.0", "id": 1, "method": "calculateSum_calculateSum", "params": [1, 2]}`),
		relayutil.GetDurationInSeconds(cf.Ingress.NATSCallWaitTimeout))

	expected := RPCCalcSumResponse{JSONRPC: "2.0", Result: 3, ID: 1}
	var actual RPCCalcSumResponse
	err = json.Unmarshal(rq.Data, &actual)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, expected, actual)
}
