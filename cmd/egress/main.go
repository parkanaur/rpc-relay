package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"rpc-relay/pkg/egress"
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

	server, err := egress.NewServer(config)
	if err != nil {
		log.Fatalln("Could not initialize egress server:", err)
	}
	server.Serve()
}
