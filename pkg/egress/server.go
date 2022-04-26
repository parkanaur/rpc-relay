package egress

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"rpc-relay/pkg/relayutil"
	"sync"
)

type Server struct {
	NATSConnection *nats.Conn
	RPCClient      *rpc.Client
	config         *relayutil.Config
	wg             *sync.WaitGroup
}

func (server *Server) Start() {
	_, err := server.NATSConnection.QueueSubscribe(
		server.config.NATS.SubjectName,
		server.config.NATS.QueueName,
		func(msg *nats.Msg) {
			log.Infoln("Incoming RPC request:", string(msg.Data))
			handleRPCRequest(&MsgContext{msg, server.RPCClient, server.config})
		})
	if err != nil {
		log.Fatalln(err)
	}
}

func (server *Server) Shutdown() error {
	if err := server.NATSConnection.Drain(); err != nil {
		return err
	}

	// waitgroup is used for NATS connection; Add() is called during server initialization and
	// Done() is called in the callback for NATS connection
	server.wg.Wait()
	log.Infoln("Stopped")
	return nil
}

type MsgContext struct {
	msg       *nats.Msg
	rpcClient *rpc.Client
	config    *relayutil.Config
}

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

	// Actual rpc call
	var result any
	err = msgCtx.rpcClient.Call(&result, rpcRequest.GetFullMethodName(), rpcRequest.Params...)
	if err != nil {
		logAndSendError(RPCErrorInternalError, msgCtx, err)
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

		err = msgCtx.msg.Respond(encodedResp)
		if err != nil {
			logAndSendError(RPCErrorInternalError, msgCtx, "Error during NATS response", err)
			return
		}
	}
}

func NewServer(config *relayutil.Config) (*Server, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)
	nc, err := nats.Connect(config.NATS.ServerURL, nats.ClosedHandler(func(_ *nats.Conn) { wg.Done() }))
	if err != nil {
		return nil, err
	}

	// Init RPC client
	url := config.JRPCServer.GetFullEndpointURL()
	rpcClient, err := rpc.DialHTTP(url)
	if err != nil {
		log.Fatalln("Could not start RPC client for", url, err)
	}
	log.Infoln("Dialed", url)

	return &Server{NATSConnection: nc, RPCClient: rpcClient, config: config, wg: &wg}, nil
}
