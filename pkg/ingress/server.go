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
	"time"
)

func NewServer(config *relayutil.Config) (http.HandlerFunc, error) {
	nc, err := nats.Connect(config.NATS.ServerURL)
	if err != nil {
		return nil, err
	}

	return func(w http.ResponseWriter, req *http.Request) {
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

		msgData, err := json.Marshal(&rpcReq)
		if err != nil {
			log.Errorln("error during NATS encoding", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		msg, err := nc.Request(
			config.NATS.GetSubjectName(rpcReq.ModuleName, rpcReq.MethodName),
			msgData,
			time.Duration(config.Ingress.NATSCallWaitTimeout*float64(time.Second)))
		if err != nil {
			log.Errorln("error during NATS RPC call", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, string(msg.Data))

	}, nil
}
