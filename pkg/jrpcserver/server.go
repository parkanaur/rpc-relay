package jrpcserver

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/parkanaur/rpc-relay/pkg/jrpcserver/services"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	log "github.com/sirupsen/logrus"
	"net/http"
)

// NewServer returns an HTTP JSON-RPC handler to plug into http.
func NewServer(config *relayutil.Config) (http.Handler, error) {
	server := rpc.NewServer()

	// Register each available service from config
	for serviceName, _ := range config.JRPCServer.EnabledRPCModules {
		if service, ok := services.ServiceRegistry[serviceName]; ok {
			err := server.RegisterName(serviceName, service())
			if err != nil {
				return nil, err
			}
			log.Infoln("Registered service module", serviceName)
		} else {
			return nil, fmt.Errorf("%v not found in service module registry", serviceName)
		}
	}

	return server, nil
}
