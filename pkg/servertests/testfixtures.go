package servertests

import (
	"context"
	gnatsd "github.com/nats-io/gnatsd/server"
	natstest "github.com/nats-io/nats-server/test"
	"github.com/parkanaur/rpc-relay/pkg/egress"
	"github.com/parkanaur/rpc-relay/pkg/jrpcserver"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	"net/http"
	"testing"
)

func NewTestConfig() *relayutil.Config {
	return &relayutil.Config{
		JRPCServer: &relayutil.JRPCServerConfig{
			Host:              "localhost",
			Port:              8001,
			RPCEndpointURL:    "/rpc",
			EnabledRPCModules: map[string][]string{"calculateSum": []string{"calculateSum"}},
			IsTLSEnabled:      false,
		},
		Ingress: &relayutil.IngressConfig{
			Host:                           "localhost",
			Port:                           8000,
			RefreshCachedRequestThreshold:  5.0,
			ExpireCachedRequestThreshold:   10.0,
			NATSCallWaitTimeout:            3.0,
			InvalidateCacheLoopSleepPeriod: 5.0,
		},
		Egress: &relayutil.EgressConfig{
			Host: "localhost",
			Port: 8002,
		},
		NATS: &relayutil.NATSConfig{
			ServerURL:   "nats://localhost:4222",
			SubjectName: "rpc.*.*",
			QueueName:   "rpcQueue",
		},
	}
}

type RelayFixture struct {
	NATSTestServer *gnatsd.Server
	JRPCServer     *http.Server
	EgressServer   *egress.Server
}

func NewJRPCServer(t *testing.T, config *relayutil.Config) *http.Server {
	srv, err := jrpcserver.NewServer(config)
	if err != nil {
		t.Fatal(err)
	}
	httpSrv := &http.Server{Addr: config.JRPCServer.GetHostWithPort(), Handler: srv}
	go func() {
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			t.Fatal(err)
		}
	}()
	return httpSrv
}

func NewRelayFixture(t *testing.T, config *relayutil.Config) *RelayFixture {
	natsSrv := natstest.RunServer(nil)
	jrpcSrv := NewJRPCServer(t, config)
	egrSrv, err := egress.NewServer(config)
	if err != nil {
		t.Fatal(err)
	}
	return &RelayFixture{natsSrv, jrpcSrv, egrSrv}
}

func (fixture *RelayFixture) Shutdown() error {
	err := fixture.EgressServer.Shutdown()
	if err != nil {
		return err
	}

	err = fixture.JRPCServer.Shutdown(context.Background())
	if err != nil {
		return err
	}

	fixture.NATSTestServer.Shutdown()
	return nil
}
