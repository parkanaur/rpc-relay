package main

import (
	"flag"
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"rpc-relay/pkg/jrpcserver"
	"rpc-relay/pkg/relayutil"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "configPath", "config.dev.json", "Config path for rpc-relay")
}

func main() {
	flag.Parse()
	config, err := relayutil.NewConfig(&configPath)
	if err != nil {
		log.Fatalln("Bad config file:", configPath, err)
	}

	server, err := jrpcserver.NewServer(config)
	if err != nil {
		log.Fatalln("Unable to start the JSON-RPC server:", err)
	}

	//http.HandleFunc(config.JRPCServer.RPCEndpointURL, server)
	addr := fmt.Sprintf("%v:%d", config.JRPCServer.Host, config.JRPCServer.Port)
	log.Infoln("Listening on", addr)
	log.Fatal(http.ListenAndServe(addr, server))
}
