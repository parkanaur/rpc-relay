package servertests

import (
	"bytes"
	"encoding/json"
	"github.com/nats-io/nats.go"
	"github.com/parkanaur/rpc-relay/pkg/egress"
	"github.com/parkanaur/rpc-relay/pkg/ingress"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestIngress_NewServer(t *testing.T) {
	cf := NewTestConfig()
	natsSrv := StartTestNATSServer(t, cf)
	defer natsSrv.Shutdown()
	srv, err := ingress.NewServer(cf)
	assert.NoError(t, err)
	defer srv.Shutdown()

	assert.NotNil(t, srv.RequestCache)
	assert.Equal(t, len(srv.RequestCache.Cache), 0)
	assert.Equal(t, srv.NATSConnection.Status(), nats.CONNECTED)
}

func TestIngress_Server_Shutdown(t *testing.T) {
	cf := NewTestConfig()
	natsSrv := StartTestNATSServer(t, cf)
	defer natsSrv.Shutdown()
	srv, err := ingress.NewServer(cf)
	assert.NoError(t, err)

	err = srv.Shutdown()
	assert.NoError(t, err)
	assert.Equal(t, srv.NATSConnection.Status(), nats.CLOSED)
}

func TestIngressHandleCall(t *testing.T) {
	cf := NewTestConfig()
	fixture := NewRelayFixture(t, cf)
	defer fixture.Shutdown()

	_, err := http.Post(
		"http://"+cf.Ingress.GetHostWithPort(),
		"application/json",
		bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 1, "method": "calculateSum_calculateSum", "params": [1, 2]}`)))
	assert.NoError(t, err)
}

func TestIngressHandleBadCalls(t *testing.T) {
	cf := NewTestConfig()
	fixture := NewRelayFixture(t, cf)
	defer fixture.Shutdown()

	data := [][]byte{
		[]byte(`{"jsonrpc": "2.0", "method": "dummyModule_dummyMethod", "params": [1,2]}`),
		[]byte(`{"id": null, "method": "dummyModule_dummyMethod", "params": [1,2]}`),
		[]byte(`{"id": 1, "jsonrpc": "1.0", "method": "dummyModule_dummyMethod", "params": [1,2]}`),
		[]byte(`{"id": 1, "jsonrpc": "2.0", "method": "dummyModuledummyMethod", "params": [1,2]}`),
		[]byte(`{"jsonrpc": "2.0", "id": 1, "method": "calculateSum_calculateSum", "params": ["1", 2]}`),
	}

	for _, v := range data {
		resp, err := http.Post(
			"http://"+cf.Ingress.GetHostWithPort(), "application/json", bytes.NewBuffer(v))
		assert.NoError(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusBadRequest)

		var response egress.RPCErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
		assert.NotNil(t, response.Error.Code)
		assert.NotNil(t, response.Error.Message)
		assert.Nil(t, response.ID)
		assert.Equal(t, response.JSONRPC, "2.0")
	}
}
