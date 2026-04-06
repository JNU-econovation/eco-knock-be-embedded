package config

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type CommonConfig struct {
	ServerHTTPPort             int
	ServerGRPCPort             int
	CentralBackendHost         string
	CentralBackendHTTPPort     int
	CentralBackendGRPCPort     int
	AllowCentralBackendFailure bool
	SensorI2CDevice            string
	SensorI2CAddress           uint8
	AirPurifierAddress         string
	AirPurifierToken           string
	AirPurifierTimeout         time.Duration
}

type yamlConfig struct {
	Server struct {
		HTTPPort int `yaml:"http_port"`
		GRPCPort int `yaml:"grpc_port"`
	} `yaml:"server"`

	CentralBackend struct {
		Host         string `yaml:"host"`
		HTTPPort     int    `yaml:"http_port"`
		GRPCPort     int    `yaml:"grpc_port"`
		AllowFailure bool   `yaml:"allow_failure"`
	} `yaml:"central_backend"`

	Sensor struct {
		I2CDevice  string `yaml:"i2c_device"`
		I2CAddress string `yaml:"i2c_address"`
	} `yaml:"sensor"`

	AirPurifier struct {
		Address string `yaml:"address"`
		Token   string `yaml:"token"`
		Timeout string `yaml:"timeout"`
	} `yaml:"air_purifier"`
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
	decoder := yaml.NewDecoder(bytes.NewReader([]byte(os.ExpandEnv(string(content)))))
	decoder.KnownFields(true)

	if err := decoder.Decode(&raw); err != nil {
		return CommonConfig{}, err
	}

	config := CommonConfig{
		ServerHTTPPort:             raw.Server.HTTPPort,
		ServerGRPCPort:             raw.Server.GRPCPort,
		CentralBackendHost:         raw.CentralBackend.Host,
		CentralBackendHTTPPort:     raw.CentralBackend.HTTPPort,
		CentralBackendGRPCPort:     raw.CentralBackend.GRPCPort,
		AllowCentralBackendFailure: raw.CentralBackend.AllowFailure,
		SensorI2CDevice:            raw.Sensor.I2CDevice,
		AirPurifierAddress:         raw.AirPurifier.Address,
		AirPurifierToken:           raw.AirPurifier.Token,
	}

	if raw.Sensor.I2CAddress != "" {
		address, err := parseI2CAddress(raw.Sensor.I2CAddress)
		if err != nil {
			return CommonConfig{}, err
		}
		config.SensorI2CAddress = address
	}

	if raw.AirPurifier.Timeout != "" {
		timeout, err := time.ParseDuration(raw.AirPurifier.Timeout)
		if err != nil {
			return CommonConfig{}, fmt.Errorf("invalid air_purifier.timeout: %w", err)
		}
		config.AirPurifierTimeout = timeout
	}

	if err := config.Validate(); err != nil {
		return CommonConfig{}, err
	}

	return config, nil
}

func (config CommonConfig) Validate() error {
	if config.ServerHTTPPort == 0 {
		return fmt.Errorf("server.http_port is required")
	}

	if config.ServerGRPCPort == 0 {
		return fmt.Errorf("server.grpc_port is required")
	}

	if config.SensorI2CDevice == "" {
		return fmt.Errorf("sensor.i2c_device is required")
	}

	if config.SensorI2CAddress == 0 {
		return fmt.Errorf("sensor.i2c_address is required")
	}

	return nil
}

func parseI2CAddress(value string) (uint8, error) {
	parsed, err := strconv.ParseUint(value, 0, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid sensor.i2c_address: %w", err)
	}

	return uint8(parsed), nil
}
