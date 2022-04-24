package main

import (
	"flag"
	"fmt"
	"github.com/nats-io/nats.go"
	log "github.com/sirupsen/logrus"
	"rpc-relay/pkg/relayutil"
	"time"
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

	n, err := nats.Connect(config.NATS.ServerURL)
	if err != nil {
		log.Fatalln(err)
	}

	msg, err := n.Request("jrpc.calculateSum.calculateSum",
		[]byte("{\"method\": \"calculateSum_calculateSum\", \"params\":[0, 2]}"),
		time.Second*3)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(string(msg.Data))
}
