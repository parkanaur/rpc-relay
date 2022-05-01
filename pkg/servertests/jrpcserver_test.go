package servertests

import (
	"bytes"
	"encoding/json"
	"github.com/parkanaur/rpc-relay/pkg/egress"
	"github.com/parkanaur/rpc-relay/pkg/jrpcserver"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func Test_JRPCServer_NewServer(t *testing.T) {
	cf := NewTestConfig()
	_, err := jrpcserver.NewServer(cf)
	assert.NoError(t, err)
}

func TestJRPCServerHandleCall(t *testing.T) {
	cf := NewTestConfig()
	fixture := NewRelayFixture(t, cf)
	defer fixture.Shutdown()

	jsonResp, err := http.Post(
		"http://"+cf.JRPCServer.GetHostWithPort(),
		"application/json",
		bytes.NewBuffer([]byte(`{"jsonrpc": "2.0", "id": 1, "method": "calculateSum_calculateSum", "params": [1, 2]}`)))
	assert.NoError(t, err)

	var resp RPCCalcSumResponse
	err = json.NewDecoder(jsonResp.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, resp.Result, 3)
	assert.Equal(t, resp.ID, 1)
	assert.Equal(t, resp.JSONRPC, "2.0")
}

func TestJRPCServerHandleBadCalls(t *testing.T) {
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
