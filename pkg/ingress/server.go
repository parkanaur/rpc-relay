package ingress

import (
	"encoding/json"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"rpc-relay/pkg/egress"
	"rpc-relay/pkg/relayutil"
	"sync"
)

// Server accepts HTTP JSON-RPC requests and proxies them to egress server via NATS
// It also holds the requets cache and runs periodic cache invalidation
type Server struct {
	// Handler for Go HTTP server
	HandlerFunc http.HandlerFunc
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

// NewServer creates a new ingress server and initializes the NATS connection
func NewServer(config *relayutil.Config) (*Server, error) {
	wg := sync.WaitGroup{}
	wg.Add(1)

	nc, err := nats.Connect(config.NATS.ServerURL, nats.ClosedHandler(func(_ *nats.Conn) { wg.Done() }))
	if err != nil {
		return nil, err
	}

	done := make(chan bool)

	reqCache := NewRequestCache()
	go reqCache.InvalidateStaleValuesLoop(config, done, &wg)

	server := &Server{nil, reqCache, nc, done, &wg, config}
	handlerFunc := func(w http.ResponseWriter, req *http.Request) {
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
		if cachedRequest, ok := reqCache.GetRequestByKey(reqKey); ok {
			var skipRenewalCheck bool
			// Check if request is expired
			if cachedRequest.IsRequestStale(
				relayutil.GetDurationInSeconds(config.Ingress.ExpireCachedRequestThreshold)) {
				err := reqCache.RemoveByKey(reqKey)
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
					relayutil.GetDurationInSeconds(config.Ingress.RefreshCachedRequestThreshold)) {
					log.Infoln("Returned cached request from cache:", reqKey)
					fmt.Fprintf(w, string(cachedRequest.response))
					return
				}
			}
		}

		msg, err := server.SendRPCRequest(rpcReq)
		if err != nil {
			log.Errorln("error during NATS RPC call", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		defer reqCache.Add(rpcReq, msg.Data)
		log.Infoln("Added request to cache:", reqKey)
		fmt.Fprintf(w, string(msg.Data))
	}

	server.HandlerFunc = handlerFunc

	return server, nil
}

// Cleanup stops the cache invalidation goroutine and drains the NATS connection
func (server *Server) Cleanup() error {
	server.wg.Add(1)
	log.Infoln("Stopping cache invalidation routine...")
	server.done <- true
	close(server.done)

	log.Infoln("Draining NATS connection...")
	if err := server.natsConn.Drain(); err != nil {
		return err
	}

	server.wg.Wait()

	return nil
}
