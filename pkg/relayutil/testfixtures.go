package relayutil

func NewTestConfig() *Config {
	return &Config{
		JRPCServer: &JRPCServerConfig{
			Host:              "localhost",
			Port:              8001,
			RPCEndpointURL:    "/rpc",
			EnabledRPCModules: map[string][]string{"calculateSum": []string{"calculateSum"}},
			IsTLSEnabled:      false,
		},
		Ingress: &IngressConfig{
			Host:                           "localhost",
			Port:                           8000,
			RefreshCachedRequestThreshold:  5.0,
			ExpireCachedRequestThreshold:   10.0,
			NATSCallWaitTimeout:            3.0,
			InvalidateCacheLoopSleepPeriod: 5.0,
		},
		Egress: &EgressConfig{
			Host: "localhost",
			Port: 8002,
		},
		NATS: &NATSConfig{
			ServerURL:   "nats://localhost:4222",
			SubjectName: "rpc.*.*",
			QueueName:   "rpcQueue",
		},
	}
}
