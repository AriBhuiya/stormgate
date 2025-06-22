package utils

import (
	"gopkg.in/yaml.v3"
	"os"
)

type Server struct {
	BindIp       string `yaml:"bind_ip"`
	BindPort     int32  `yaml:"bind_port"`
	ReadTimeOut  int64  `yaml:"read_time_out"`
	WriteTimeOut int64  `yaml:"write_time_out"`
}

type Balancer struct {
	RoutingStrategy string `yaml:"routing_strategy"`
}

type Services struct {
	Name       string   `yaml:"name"`
	PathPrefix string   `yaml:"path_prefix"`
	Strategy   string   `yaml:"strategy"`
	Backends   []string `yaml:"backends"`
}

type Config struct {
	Server   Server     `yaml:"server"`
	Services []Services `yaml:"services"`
	Balancer Balancer   `yaml:"balancer"`
}

func LoadConfig(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
