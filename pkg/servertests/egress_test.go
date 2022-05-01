package servertests

import (
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/parkanaur/rpc-relay/pkg/egress"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	"github.com/stretchr/testify/assert"
	"testing"
)

type RPCCalcSumResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  int    `json:"result"`
	ID      int    `json:"id"`
}

func TestEgress_NewServer(t *testing.T) {
	cf := NewTestConfig()
	natsSrv := StartTestNATSServer(t, cf)
	defer natsSrv.Shutdown()
	server, err := egress.NewServer(cf)
	assert.NoError(t, err)

	assert.NotNil(t, server.NATSConnection)
	assert.Equal(t, server.NATSConnection.ConnectedUrl(), cf.NATS.ServerURL)
	assert.Equal(t, server.NATSConnection.Status(), nats.CONNECTED)

	assert.NotNil(t, server.RPCClient)

	_ = server.Shutdown()
}

func TestEgress_NewServer_Shutdown(t *testing.T) {
	cf := NewTestConfig()
	natsSrv := StartTestNATSServer(t, cf)
	defer natsSrv.Shutdown()
	server, _ := egress.NewServer(cf)

	err := server.Shutdown()
	assert.NoError(t, err)

	assert.Equal(t, server.NATSConnection.Status(), nats.CLOSED)
}

func TestEgress_Server_handleRPCRequest(t *testing.T) {
	cf := NewTestConfig()
	fixture := NewRelayFixture(t, cf)
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
