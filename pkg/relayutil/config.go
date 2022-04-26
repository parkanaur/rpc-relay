package relayutil

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// JRPCServerConfig is a part of the config which holds config values for the JSONRPC server
type JRPCServerConfig struct {
	Host              string
	Port              int
	RPCEndpointURL    string
	EnabledRPCModules map[string][]string
	IsTLSEnabled      bool
}

// GetFullEndpointURL generates a full HTTP URL for JSON-RPC endpoint from the config
func (config *JRPCServerConfig) GetFullEndpointURL() string {
	protocol := "http"
	if config.IsTLSEnabled {
		protocol = "https"
	}
	return fmt.Sprintf("%v://%v:%d%v", protocol, config.Host, config.Port, config.RPCEndpointURL)
}

// GetHostWithPort Returns a host:port for JSON-RPC server
func (config *JRPCServerConfig) GetHostWithPort() string {
	return fmt.Sprintf("%v:%d", config.Host, config.Port)
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
	// Timeout for NATS/RPC call to egress proxy, after which HTTP 504 is returned
	NATSCallWaitTimeout float64
	// Run cache invalidation each N seconds
	InvalidateCacheLoopSleepPeriod float64
}

// GetHostWithPort Returns a host:port for the ingress server
func (config *IngressConfig) GetHostWithPort() string {
	return fmt.Sprintf("%v:%d", config.Host, config.Port)
}

// EgressConfig is a part of the config which holds config values for the egress proxy
type EgressConfig struct {
	Host string
	Port int
}

// NATSConfig is a part of the config which holds config values for the NATS server
type NATSConfig struct {
	ServerURL   string
	SubjectName string
	QueueName   string
}

// GetSubjectName returns a full NATS subject given RPC call's method/module names
func (config *NATSConfig) GetSubjectName(moduleName, methodName string) string {
	subj := strings.Replace(config.SubjectName, "*", moduleName, 1)
	subj = strings.Replace(subj, "*", methodName, 1)
	return subj
}

// Config is a struct for holding configuration values for all proxies and servers
type Config struct {
	JRPCServer JRPCServerConfig
	Ingress    IngressConfig
	Egress     EgressConfig
	NATS       NATSConfig
}

// GetDurationInSeconds converts a float value for seconds to time.Duration
func GetDurationInSeconds(configSecondsValue float64) time.Duration {
	return time.Duration(configSecondsValue * float64(time.Second))
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
