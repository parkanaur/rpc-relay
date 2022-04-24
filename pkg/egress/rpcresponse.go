package egress

import "fmt"

type RPCResponse struct {
	JSONRPC string
	Result  any
	ID      any
}

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

var errorMap = map[RPCErrorNum]string{
	RPCErrorNotWellFormed:    "not well formed",
	RPCErrorInvalidRequest:   "invalid request",
	RPCErrorMethodNotFound:   "method not found",
	RPCErrorInvalidParams:    "invalid params",
	RPCErrorInternalError:    "internal error",
	RPCErrorModuleNotEnabled: "module not enabled",
}

type RPCError struct {
	Code    RPCErrorNum
	Message string
}

type RPCErrorResponse struct {
	JSONRPC string
	ID      any
	Error   *RPCError
}

func CreateResponse(result any, request *RPCRequest) *RPCResponse {
	return &RPCResponse{
		Result:  result,
		ID:      request.ID,
		JSONRPC: request.JSONRPC,
	}
}

func CreateErrorResponse(num RPCErrorNum, info ...any) *RPCErrorResponse {
	return &RPCErrorResponse{
		JSONRPC: "2.0",
		ID:      nil,
		Error:   &RPCError{num, errorMap[num] + fmt.Sprintln(info...)},
	}
}
