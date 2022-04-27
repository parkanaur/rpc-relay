package main

import (
	"context"
	"flag"
	"github.com/parkanaur/rpc-relay/pkg/jrpcserver"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	log "github.com/sirupsen/logrus"
	"net/http"
	"os"
	"os/signal"
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

	server, err := jrpcserver.NewServer(config)
	if err != nil {
		log.Fatalln("Unable to start the JSON-RPC server:", err)
	}

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	httpServer := &http.Server{Addr: config.JRPCServer.GetHostWithPort(), Handler: server}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("Error while serving HTTP:", err)
		}
	}()
	log.Infoln("Listening on", httpServer.Addr)

	<-done
	log.Infoln("Stopping...")
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
	}()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalln("Server shutdown failed:", err)
	}
}
