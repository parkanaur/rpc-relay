package egress

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/nats-io/nats.go"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	log "github.com/sirupsen/logrus"
	"strings"
	"sync"
)

// Server is a structure for the egress server holding the NATS listener and a client for JSON-RPC calls
type Server struct {
	// NATS listener. These are launched in an RPC queue (see config)
	NATSConnection *nats.Conn
	// Client for sending correct requests to the JSON-RPC server
	RPCClient *rpc.Client
	// Server config
	config *relayutil.Config
	// Used during draining of the NATS connection
	wg *sync.WaitGroup
}

// Shutdown drains the NATS connection and closes the RPC client
func (server *Server) Shutdown() error {
	if err := server.NATSConnection.Drain(); err != nil {
		return err
	}

	server.RPCClient.Close()
	// waitgroup is used for NATS connection; Add() is called during server initialization and
	// Done() is called in the callback for NATS connection
	server.wg.Wait()
	log.Infoln("Stopped")
	return nil
}

// MsgContext is an auxiliary structure for passing around certain useful variables
type MsgContext struct {
	msg       *nats.Msg
	rpcClient *rpc.Client
	config    *relayutil.Config
}

// logAndSendError logs the error to stderr and returns an RPCErrorResponse to the ingress server
func logAndSendError(errNum RPCErrorNum, msgCtx *MsgContext, info ...any) {
	log.Errorln(info...)
	// Info is prevented from being returned to user on purpose to avoid disclosing sensitive
	// error info
	resp, err := json.Marshal(CreateErrorResponse(errNum))
	if err != nil {
		log.Errorln("Error while marshalling error:", err)
		return
	}

	err = msgCtx.msg.Respond(resp)
	if err != nil {
		log.Errorln("Error during NATS response", err)
		return
	}
}

// handleRPCRequest handles incoming NATS messages, sends RPC requests and replies to NATS messages
func handleRPCRequest(msgCtx *MsgContext) {
	rpcRequest, err := ParseCall(msgCtx.msg.Data)
	if err != nil {
		logAndSendError(RPCErrorNotWellFormed, msgCtx, string(msgCtx.msg.Data), err)
		return
	}
	// Checking if method is available for calling
	if _, ok := msgCtx.config.JRPCServer.EnabledRPCModules[rpcRequest.ModuleName]; !ok {
		logAndSendError(RPCErrorModuleNotEnabled, msgCtx, rpcRequest.ModuleName)
		return
	}
	// TODO: implement check for whether the method is enabled (via EnabledRPCMethods dict values for
	// given module). Possibly requires rewriting the method list to be a map for
	// faster checks.

	// Actual rpc call
	var result any
	err = msgCtx.rpcClient.Call(&result, rpcRequest.GetFullMethodName(), rpcRequest.Params...)
	if err != nil {
		errStr := err.Error()
		var rpcErrNum RPCErrorNum = RPCErrorInternalError

		// Filter out errors caused by user's incorrect requests
		for errorPrefix, errorNum := range RPCErrorMap {
			if strings.HasPrefix(errStr, errorPrefix) {
				rpcErrNum = errorNum
			}
		}
		logAndSendError(rpcErrNum, msgCtx, err)
		return
	}

	// Encoding and sending the result
	if result != nil {
		resp := CreateResponse(result, rpcRequest)
		encodedResp, err := json.Marshal(&resp)
		if err != nil {
			logAndSendError(RPCErrorInternalError, msgCtx, "Error during JSON response encoding", err)
			return
		}

		// NATS reply
		err = msgCtx.msg.Respond(encodedResp)
		if err != nil {
			logAndSendError(RPCErrorInternalError, msgCtx, "Error during NATS response", err)
			return
		}
	}
}

// NewServer creates a new egress server from the config
func NewServer(config *relayutil.Config) (*Server, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	// Init NATS
	nc, err := nats.Connect(config.NATS.ServerURL, nats.ClosedHandler(func(_ *nats.Conn) { wg.Done() }))
	if err != nil {
		return nil, err
	}

	// Init RPC client
	url := config.JRPCServer.GetFullEndpointURL()
	rpcClient, err := rpc.DialHTTP(url)
	if err != nil {
		return nil, err
	}
	log.Infoln("Dialed", url)

	_, err = nc.QueueSubscribe(
		config.NATS.SubjectName,
		config.NATS.QueueName,
		func(msg *nats.Msg) {
			log.Infoln("Incoming RPC request:", string(msg.Data))
			go handleRPCRequest(&MsgContext{msg, rpcClient, config})
		})
	if err != nil {
		return nil, err
	}

	return &Server{NATSConnection: nc, RPCClient: rpcClient, config: config, wg: &wg}, nil
}
