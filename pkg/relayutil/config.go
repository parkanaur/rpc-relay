package relayutil

import (
	"encoding/json"
	"os"
)

// JRPCServerConfig is a part of the config which holds config values for the JSONRPC server
type JRPCServerConfig struct {
	Host              string
	Port              int
	EnabledRPCMethods []string
}

// IngressConfig is a part of the config which holds config values for the ingress proxy
type IngressConfig struct {
	Host string
	Port int
	// Threshold value in seconds. If less time than threshold had passed before a cached request was retrieved,
	// a new request is not made (the cached value is still immediately returned either way)
	RefreshCachedRequestThreshold float64
	// Threshold value in seconds. All cached requests are expired if they're stored in the cache for longer
	// than threshold value
	ExpireCachedRequestThreshold float64
}

// EgressConfig is a part of the config which holds config values for the egress proxy
type EgressConfig struct {
	Host string
	Port int
}

// NATSConfig is a part of the config which holds config values for the NATS server
type NATSConfig struct {
	ServerURL string
}

// Config is a struct for holding configuration values for all proxies and servers
type Config struct {
	JRPCServer JRPCServerConfig
	Ingress    IngressConfig
	Egress     EgressConfig
	NATS       NATSConfig
}

// Parse takes a JSON file at configPath and attempts to parse its contents into the config struct
func (config *Config) Parse(configPath *string) error {
	jsonFile, err := os.Open(*configPath)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	jsonDecoder := json.NewDecoder(jsonFile)
	err = jsonDecoder.Decode(config)
	if err != nil {
		return err
	}

	return nil
}

// NewConfig generates a config struct and parses a config file into it
func NewConfig(configPath *string) (*Config, error) {
	config := new(Config)
	if err := config.Parse(configPath); err != nil {
		return nil, err
	}
	return config, nil
}
