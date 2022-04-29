package ingress

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	"github.com/parkanaur/rpc-relay/pkg/egress"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"sync"
)

// Server accepts HTTP JSON-RPC requests and proxies them to egress server via NATS
// It also holds the requets cache and runs periodic cache invalidation
type Server struct {
	// RPC request cache
	RequestCache *RequestCache
	// NATS connection
	natsConn *nats.Conn
	// Channel which is written to during shutdown and read from by the shutdown function
	done chan bool
	// Waitgroup for NATS connection draining handling
	wg *sync.WaitGroup
	// Server config
	config *relayutil.Config
}

// SendRPCRequest creates a NATS request to egress and returns the NATS reply
func (server *Server) SendRPCRequest(request *egress.RPCRequest) (*nats.Msg, error) {
	msgData, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}

	return server.natsConn.Request(
		server.config.NATS.GetSubjectName(request.ModuleName, request.MethodName),
		msgData,
		relayutil.GetDurationInSeconds(server.config.Ingress.NATSCallWaitTimeout))
}

func (server *Server) Handler(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "invalid HTTP method: only POST is allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorln("error during body reading", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	rpcReq, err := egress.ParseCall(body)
	if err != nil {
		http.Error(w, fmt.Sprintln("invalid JSON format:", err), http.StatusBadRequest)
		return
	}

	reqKey := rpcReq.GetRequestKey()
	if cachedRequest, ok := server.RequestCache.GetRequestByKey(reqKey); ok {
		var skipRenewalCheck bool
		// Check if request is expired
		if cachedRequest.IsRequestStale(
			relayutil.GetDurationInSeconds(server.config.Ingress.ExpireCachedRequestThreshold)) {
			err := server.RequestCache.RemoveByKey(reqKey)
			if err != nil {
				log.Errorln("Failed to remove by key", reqKey, err)
			}
			skipRenewalCheck = true
		}

		// Check if request has to be renewed after the cached value is returned.
		// Return request immediately if it's fresh enough.
		//
		// The assumption here is that a request older than some value but young enough not to be expired
		// has to be renewed and the new result is returned afterwards.
		// It is also possible to return the old result and then defer SendRPCRequest to renew the result
		// after the user has already gotten their old result, but I assume this is not what was required.
		if !skipRenewalCheck {
			if !cachedRequest.IsRequestStale(
				relayutil.GetDurationInSeconds(server.config.Ingress.RefreshCachedRequestThreshold)) {
				log.Infoln("Returned cached request from cache:", reqKey)
				w.Write(cachedRequest.response)
				return
			}
		}
	}

	// TODO: Check if response is an ErrorResponse AND the error code is for an internal error.
	// Return a http.Error with HTTP 500 in this case. Forward the error RPC response as usual otherwise.

	msg, err := server.SendRPCRequest(rpcReq)
	if err != nil {
		log.Errorln("error during NATS RPC call", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	defer server.RequestCache.Add(rpcReq, msg.Data)
	log.Infoln("Added request to cache:", reqKey)
	w.Write(msg.Data)
}

// NewServer creates a new ingress server and initializes the NATS connection
func NewServer(config *relayutil.Config) (*Server, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	nc, err := nats.Connect(config.NATS.ServerURL, nats.ClosedHandler(func(_ *nats.Conn) { wg.Done() }))
	if err != nil {
		return nil, err
	}

	done := make(chan bool)

	reqCache := NewRequestCache(config)
	reqCache.Start()

	server := &Server{reqCache, nc, done, &wg, config}

	return server, nil
}

// Cleanup stops the cache invalidation goroutine and drains the NATS connection
func (server *Server) Cleanup() error {
	log.Infoln("Stopping cache invalidation routine...")
	server.RequestCache.Stop()

	log.Infoln("Draining NATS connection...")
	if err := server.natsConn.Drain(); err != nil {
		return err
	}

	server.wg.Wait()

	return nil
}
