package egress

import "fmt"

// RPCResponse holds the JSON-RPC 2.0 spec response object, provided the request/response
// were without errors
type RPCResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  any    `json:"result"`
	ID      any    `json:"id"`
}

// RPCErrorNum is an alias to help distinguish between RPC spec error codes
type RPCErrorNum int

// TODO: Add more codes or use them from elsewhere (geth's rpc codes are not exported)
const (
	RPCErrorNotWellFormed    RPCErrorNum = -32700
	RPCErrorInvalidRequest               = -32600
	RPCErrorMethodNotFound               = -32601
	RPCErrorInvalidParams                = -32602
	RPCErrorInternalError                = -32603
	RPCErrorModuleNotEnabled             = 101
)

const (
	RPCPrefixInvalidArgument string = "invalid argument"
)

// RPCErrorMap is used for checking responses from jrpcserver/geth JSONRPC server
var RPCErrorMap = map[string]RPCErrorNum{
	RPCPrefixInvalidArgument: RPCErrorInvalidParams,
}

// Used for responding to ingress server
var errorResponseMap = map[RPCErrorNum]string{
	RPCErrorNotWellFormed:    "not well formed",
	RPCErrorInvalidRequest:   "invalid request",
	RPCErrorMethodNotFound:   "method not found",
	RPCErrorInvalidParams:    "invalid params",
	RPCErrorInternalError:    "internal error",
	RPCErrorModuleNotEnabled: "module not enabled",
}

// RPCError is a JSON-RPC 2.0 error response field
type RPCError struct {
	Code    RPCErrorNum `json:"code"`
	Message string      `json:"message"`
}

// RPCErrorResponse is a JSON-RPC 2.0 response object that is returned instead of RPCResponse
// if there were errors in the request/during the call
type RPCErrorResponse struct {
	JSONRPC string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Error   *RPCError `json:"error"`
}

// CreateResponse serializes the method call's result into a RPCResponse
func CreateResponse(result any, request *RPCRequest) *RPCResponse {
	return &RPCResponse{
		Result:  result,
		ID:      request.ID,
		JSONRPC: request.JSONRPC,
	}
}

// CreateErrorResponse serializes an error into a RPCErrorResponse
func CreateErrorResponse(num RPCErrorNum, info ...any) *RPCErrorResponse {
	return &RPCErrorResponse{
		JSONRPC: "2.0",
		ID:      nil,
		Error:   &RPCError{num, errorResponseMap[num] + fmt.Sprintln(info...)},
	}
}
