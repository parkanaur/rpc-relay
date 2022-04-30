package egress

import (
	"fmt"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	log "github.com/sirupsen/logrus"
	"testing"
)

func NewTestConfig() *relayutil.Config {
	config, err := relayutil.NewConfig("../../config.test.json")
	if err != nil {
		log.Fatalln("bad config file", "config.test.json")
	}
	return config
}

func TestNewServer(t *testing.T) {
	cf := NewTestConfig()
	fmt.Println(cf)
}
