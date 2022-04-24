package egress

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RPCRequest struct {
	moduleName string
	methodName string

	// JSONRPC spec fields
	Params  []any
	ID      any
	Method  string
	JSONRPC string
}

func (call *RPCRequest) GetFullMethodName() string {
	return fmt.Sprintf("%v_%v", call.methodName, call.moduleName)
}

func ParseCall(data []byte) (*RPCRequest, error) {
	var call RPCRequest
	if err := json.Unmarshal(data, &call); err != nil {
		return nil, err
	}
	// JSONRPC specific checks
	if call.ID == nil {
		return nil, fmt.Errorf("missing ID field")
	}
	if call.JSONRPC != "2.0" {
		return nil, fmt.Errorf("bad jsonrpc field")
	}

	// go-ethereum's JSONRPC 2.0 implementation creates methods using the `service_method` name template
	// Go's standard RPC lib is 1.0-only
	s := strings.Split(call.Method, "_")
	if len(s) != 2 || s[0] == "" || s[1] == "" {
		return nil, fmt.Errorf("bad RPC call: %v", call.Method)
	}
	call.moduleName = s[0]
	call.methodName = s[1]

	return &call, nil
}
