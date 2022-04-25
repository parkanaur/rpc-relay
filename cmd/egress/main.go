package main

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"rpc-relay/pkg/egress"
	"rpc-relay/pkg/relayutil"
	"syscall"
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

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	server, err := egress.NewServer(config)
	if err != nil {
		log.Fatalln("Could not initialize egress server:", err)
	}

	go func() {
		server.Serve()
	}()
	<-done
	log.Infoln("Stopping...")
	if err := server.Shutdown(); err != nil {
		log.Fatalln("Error while shutting down:", err)
	}
}
