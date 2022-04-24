package main

import (
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/rpc"
	log "github.com/sirupsen/logrus"
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

	url := config.JRPCServer.GetFullEndpointURL()
	client, err := rpc.DialHTTP(url)
	if err != nil {
		log.Fatalln("Could not start RPC client", configPath, err)
	}
	log.Infoln("Dialed", url)

	var result int
	err = client.Call(&result, "calculateSum_calculateSum", 1, 2)
	if err != nil {
		log.Fatalln("Error calling RPC service", err)
	}
	fmt.Println(result)
}
