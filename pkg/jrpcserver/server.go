package jrpcserver

import (
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"rpc-relay/pkg/jrpcserver/services"
	"rpc-relay/pkg/relayutil"
)

// HTTPConn is an auxiliary structure for wrapping RPC calls into HTTP requests
type HTTPConn struct {
	in  io.Reader
	out io.Writer
}

// Read
func (c *HTTPConn) Read(p []byte) (n int, err error)  { return c.in.Read(p) }
func (c *HTTPConn) Write(d []byte) (n int, err error) { return c.out.Write(d) }
func (c *HTTPConn) Close() error                      { return nil }

// JRPCServerHandler is an JSON-RPC server handler to be used as an argument to http.HandleFunc
// type JRPCServerHandler func(w http.ResponseWriter, r *http.Request) error

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
