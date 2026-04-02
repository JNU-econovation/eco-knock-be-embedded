package config

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type CommonConfig struct {
	ServerHTTPPort             int
	CentralBackendHost         string
	CentralBackendHTTPPort     int
	CentralBackendGRPCPort     int
	AllowCentralBackendFailure bool
}

type yamlConfig struct {
	Server struct {
		HTTPPort int `yaml:"http_port"`
	} `yaml:"server"`

	CentralBackend struct {
		Host         string `yaml:"host"`
		HTTPPort     int    `yaml:"http_port"`
		GRPCPort     int    `yaml:"grpc_port"`
		AllowFailure bool   `yaml:"allow_failure"`
	} `yaml:"central_backend"`
}

func MustLoad(path string) CommonConfig {
	config, err := Load(path)
	if err != nil {
		log.Fatal(err)
	}

	return config
}

func Load(path string) (CommonConfig, error) {
	if path == "" {
		return CommonConfig{}, fmt.Errorf("config path is required")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return CommonConfig{}, err
	}

	raw := yamlConfig{}
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	decoder.KnownFields(true)

	if err := decoder.Decode(&raw); err != nil {
		return CommonConfig{}, err
	}

	config := CommonConfig{
		ServerHTTPPort:             raw.Server.HTTPPort,
		CentralBackendHost:         raw.CentralBackend.Host,
		CentralBackendHTTPPort:     raw.CentralBackend.HTTPPort,
		CentralBackendGRPCPort:     raw.CentralBackend.GRPCPort,
		AllowCentralBackendFailure: raw.CentralBackend.AllowFailure,
	}

	if err := config.Validate(); err != nil {
		return CommonConfig{}, err
	}

	return config, nil
}

func (config CommonConfig) Validate() error {
	if config.ServerHTTPPort == 0 {
		return fmt.Errorf("server.http_port 설정이 필요합니다")
	}

	if config.AllowCentralBackendFailure {
		return nil
	}

	if config.CentralBackendHost == "" {
		return fmt.Errorf("central_backend.host 설정이 필요합니다")
	}

	if config.CentralBackendHTTPPort == 0 {
		return fmt.Errorf("central_backend.http_port 설정이 필요합니다")
	}

	if config.CentralBackendGRPCPort == 0 {
		return fmt.Errorf("central_backend.grpc_port 설정이 필요합니다")
	}

	return nil
}
