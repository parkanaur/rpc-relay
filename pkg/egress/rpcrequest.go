package egress

import (
	"encoding/json"
	"fmt"
	"strings"
)

// RPCRequest holds the incoming JSON-RPC 2.0 request data
type RPCRequest struct {
	// First part of the method name ("calculateSum1_calculateSum") -> "calculateSum1"
	ModuleName string `json:"-"`
	// Second part of the method name ("calculateSum1_calculateSum") -> "calculateSum"
	MethodName string `json:"-"`

	// JSONRPC spec fields
	Params  []any  `json:"params"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	JSONRPC string `json:"jsonrpc"`
}

// GetRequestKey returns a string to be used as a request cache key. The key uniquely identifies
// the request by its method name and parameters
// calculateSum(1,2) is different from calculateSum(2,1)
func (call *RPCRequest) GetRequestKey() string {
	return fmt.Sprintln(call.Method, call.Params)
}

// GetFullMethodName forms a full method name from its module/method parts
func (call *RPCRequest) GetFullMethodName() string {
	return fmt.Sprintf("%v_%v", call.MethodName, call.ModuleName)
}

// ParseCall serializes an incoming RPC request into the actual RPCRequest object
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
