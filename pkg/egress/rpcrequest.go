package egress

import (
	"encoding/json"
	"fmt"
	"strings"
)

type RPCRequest struct {
	ModuleName string `json:"-"`
	MethodName string `json:"-"`

	// JSONRPC spec fields
	Params  []any  `json:"params"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	JSONRPC string `json:"jsonrpc"`
}

func (call *RPCRequest) GetFullMethodName() string {
	return fmt.Sprintf("%v_%v", call.MethodName, call.ModuleName)
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
	call.ModuleName = s[0]
	call.MethodName = s[1]

	// TODO: Param checking for a given method based on types defined in config

	return &call, nil
}
