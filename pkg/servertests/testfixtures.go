package servertests

import (
	"context"
	gnatsd "github.com/nats-io/gnatsd/server"
	natstest "github.com/nats-io/nats-server/test"
	"github.com/parkanaur/rpc-relay/pkg/egress"
	"github.com/parkanaur/rpc-relay/pkg/ingress"
	"github.com/parkanaur/rpc-relay/pkg/jrpcserver"
	"github.com/parkanaur/rpc-relay/pkg/relayutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"testing"
)

func StartTestNATSServer(t *testing.T, cf *relayutil.Config) *gnatsd.Server {
	u, err := url.ParseRequestURI(cf.NATS.ServerURL)
	host, portStr, err := net.SplitHostPort(u.Host)
	port, err := strconv.Atoi(portStr)
	opts := gnatsd.Options{
		Host:           host,
		Port:           port,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
	}
	if err != nil {
		t.Fatal(err)
	}

	return natstest.RunServer(&opts)
}

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
			ServerURL:   "nats://localhost:4223",
			SubjectName: "rpc.*.*",
			QueueName:   "rpcQueue",
		},
	}
}

type RelayFixture struct {
	NATSTestServer *gnatsd.Server
	JRPCServer     *http.Server
	EgressServer   *egress.Server
	IngressSever   *http.Server
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

func NewIngressServer(t *testing.T, config *relayutil.Config) *http.Server {
	srv, err := ingress.NewServer(config)
	if err != nil {
		t.Fatal(err)
	}
	httpServer := &http.Server{Addr: config.Ingress.GetHostWithPort(), Handler: srv}

	return httpServer
}

func NewRelayFixture(t *testing.T, config *relayutil.Config) *RelayFixture {
	natsSrv := StartTestNATSServer(t, config)
	jrpcSrv := NewJRPCServer(t, config)
	egrSrv, err := egress.NewServer(config)
	if err != nil {
		t.Fatal(err)
	}
	ingSrv := NewIngressServer(t, config)
	return &RelayFixture{natsSrv, jrpcSrv, egrSrv, ingSrv}
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
