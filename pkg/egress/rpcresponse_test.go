package egress

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type DummyResult struct {
	rid string
	ret float64
}

func TestCreateResponse(t *testing.T) {
	results := append(make([]any, 0), 1, "strRes", 10.45, &DummyResult{"12", 34.56}, nil)
	req := NewDummyRPCRequest()
	for _, result := range results {
		resp := CreateResponse(result, req)
		assert.Equal(t, resp.JSONRPC, "2.0")
		assert.Equal(t, resp.ID, req.ID)
		assert.Equal(t, resp.Result, result)
	}
}

func TestCreateErrorResponse(t *testing.T) {
	for errNum, errMsg := range errorResponseMap {
		resp := CreateErrorResponse(errNum)
		assert.Equal(t, resp.JSONRPC, "2.0")
		assert.Nil(t, resp.ID)
		assert.Equal(t, resp.Error.Code, errNum)
		assert.Equal(t, resp.Error.Message, errMsg)
	}
}

func TestCreateErrorResponseWithCustomInfo(t *testing.T) {
	info := append(make([]any, 0), "err1", "err2")
	resp := CreateErrorResponse(RPCErrorInvalidParams, info...)
	assert.Equal(t, resp.JSONRPC, "2.0")
	assert.Nil(t, resp.ID)
	assert.Equal(t, int(resp.Error.Code), RPCErrorInvalidParams)
	assert.Equal(t, resp.Error.Message, fmt.Sprintf("%v %v %v", errorResponseMap[RPCErrorInvalidParams], info[0], info[1]))
}
