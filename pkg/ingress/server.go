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
)

type Server struct {
	HandlerFunc  http.HandlerFunc
	RequestCache *RequestCache
	Done         chan bool
}

func SendRPCRequest(request *egress.RPCRequest, nc *nats.Conn, config *relayutil.Config) (*nats.Msg, error) {
	msgData, err := json.Marshal(&request)
	if err != nil {
		return nil, err
	}

	return nc.Request(
		config.NATS.GetSubjectName(request.ModuleName, request.MethodName),
		msgData,
		relayutil.GetDurationInSeconds(config.Ingress.NATSCallWaitTimeout))
}

func NewServer(config *relayutil.Config) (*Server, error) {
	nc, err := nats.Connect(config.NATS.ServerURL)
	if err != nil {
		return nil, err
	}

	done := make(chan bool)

	reqCache := NewRequestCache()
	go reqCache.InvalidateStaleValuesLoop(config, done)

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
			http.Error(w, "invalid JSON format", http.StatusBadRequest)
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
			if !skipRenewalCheck {
				if !cachedRequest.IsRequestStale(
					relayutil.GetDurationInSeconds(config.Ingress.RefreshCachedRequestThreshold)) {
					fmt.Fprintf(w, string(cachedRequest.response))
					return
				}
			}
		}

		msg, err := SendRPCRequest(rpcReq, nc, config)
		if err != nil {
			log.Errorln("error during NATS RPC call", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		defer reqCache.Add(rpcReq, msg.Data)
		fmt.Fprintf(w, string(msg.Data))
	}

	return &Server{handlerFunc, reqCache, done}, nil
}
