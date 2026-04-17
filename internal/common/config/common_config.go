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
	ServerHTTPPort                               int
	ServerGRPCPort                               int
	CentralBackendHost                           string
	CentralBackendHTTPPort                       int
	CentralBackendGRPCPort                       int
	AllowCentralBackendFailure                   bool
	SensorI2CDevice                              string
	SensorI2CAddress                             uint8
	SensorStateDBPath                            string
	SensorHeaterTempC                            uint16
	SensorHeaterDuration                         time.Duration
	SensorAmbientTempC                           int8
	SensorPollInterval                           time.Duration
	SensorStateCheckpointValidSamples            int
	SensorAirQualityHistoryLimit                 int
	SensorAirQualityStabilizationDuration        time.Duration
	SensorAirQualityLearningDuration             time.Duration
	SensorAirQualityStabilizationValidSampleGoal int
	SensorAirQualityLearningValidSampleGoal      int
	AirPurifierAddress                           string
	AirPurifierToken                             string
	AirPurifierTimeout                           time.Duration
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
		I2CDevice                   string `yaml:"i2c_device"`
		I2CAddress                  string `yaml:"i2c_address"`
		StateDBPath                 string `yaml:"state_db_path"`
		HeaterTempC                 uint16 `yaml:"heater_temp_c"`
		HeaterDuration              string `yaml:"heater_duration"`
		AmbientTempC                int8   `yaml:"ambient_temp_c"`
		PollInterval                string `yaml:"poll_interval"`
		StateCheckpointValidSamples int    `yaml:"state_checkpoint_valid_samples"`
		AirQuality                  struct {
			HistoryLimit                 int    `yaml:"history_limit"`
			StabilizationDuration        string `yaml:"stabilization_duration"`
			LearningDuration             string `yaml:"learning_duration"`
			StabilizationValidSampleGoal int    `yaml:"stabilization_valid_sample_goal"`
			LearningValidSampleGoal      int    `yaml:"learning_valid_sample_goal"`
		} `yaml:"air_quality"`
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
		log.Fatalf("애플리케이션 설정을 불러오지 못했습니다: %v", err)
	}

	return config
}

func Load(path string) (CommonConfig, error) {
	if path == "" {
		return CommonConfig{}, fmt.Errorf("설정 파일 경로가 필요합니다")
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
		ServerHTTPPort:                               raw.Server.HTTPPort,
		ServerGRPCPort:                               raw.Server.GRPCPort,
		CentralBackendHost:                           raw.CentralBackend.Host,
		CentralBackendHTTPPort:                       raw.CentralBackend.HTTPPort,
		CentralBackendGRPCPort:                       raw.CentralBackend.GRPCPort,
		AllowCentralBackendFailure:                   raw.CentralBackend.AllowFailure,
		SensorI2CDevice:                              raw.Sensor.I2CDevice,
		SensorStateDBPath:                            raw.Sensor.StateDBPath,
		SensorHeaterTempC:                            raw.Sensor.HeaterTempC,
		SensorAmbientTempC:                           raw.Sensor.AmbientTempC,
		SensorStateCheckpointValidSamples:            raw.Sensor.StateCheckpointValidSamples,
		SensorAirQualityHistoryLimit:                 raw.Sensor.AirQuality.HistoryLimit,
		SensorAirQualityStabilizationValidSampleGoal: raw.Sensor.AirQuality.StabilizationValidSampleGoal,
		SensorAirQualityLearningValidSampleGoal:      raw.Sensor.AirQuality.LearningValidSampleGoal,
		AirPurifierAddress:                           raw.AirPurifier.Address,
		AirPurifierToken:                             raw.AirPurifier.Token,
	}

	if raw.Sensor.I2CAddress != "" {
		address, err := parseI2CAddress(raw.Sensor.I2CAddress)
		if err != nil {
			return CommonConfig{}, err
		}
		config.SensorI2CAddress = address
	}

	if raw.Sensor.HeaterDuration != "" {
		duration, err := time.ParseDuration(raw.Sensor.HeaterDuration)
		if err != nil {
			return CommonConfig{}, fmt.Errorf("sensor.heater_duration 값이 올바르지 않습니다: %w", err)
		}
		config.SensorHeaterDuration = duration
	}

	if raw.Sensor.PollInterval != "" {
		duration, err := time.ParseDuration(raw.Sensor.PollInterval)
		if err != nil {
			return CommonConfig{}, fmt.Errorf("sensor.poll_interval 값이 올바르지 않습니다: %w", err)
		}
		config.SensorPollInterval = duration
	}

	if raw.Sensor.AirQuality.StabilizationDuration != "" {
		duration, err := time.ParseDuration(raw.Sensor.AirQuality.StabilizationDuration)
		if err != nil {
			return CommonConfig{}, fmt.Errorf("sensor.air_quality.stabilization_duration 값이 올바르지 않습니다: %w", err)
		}
		config.SensorAirQualityStabilizationDuration = duration
	}

	if raw.Sensor.AirQuality.LearningDuration != "" {
		duration, err := time.ParseDuration(raw.Sensor.AirQuality.LearningDuration)
		if err != nil {
			return CommonConfig{}, fmt.Errorf("sensor.air_quality.learning_duration 값이 올바르지 않습니다: %w", err)
		}
		config.SensorAirQualityLearningDuration = duration
	}

	if raw.AirPurifier.Timeout != "" {
		timeout, err := time.ParseDuration(raw.AirPurifier.Timeout)
		if err != nil {
			return CommonConfig{}, fmt.Errorf("air_purifier.timeout 값이 올바르지 않습니다: %w", err)
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
		return fmt.Errorf("server.http_port 값이 필요합니다")
	}

	if config.ServerGRPCPort == 0 {
		return fmt.Errorf("server.grpc_port 값이 필요합니다")
	}

	if config.SensorI2CDevice == "" {
		return fmt.Errorf("sensor.i2c_device 값이 필요합니다")
	}

	if config.SensorI2CAddress == 0 {
		return fmt.Errorf("sensor.i2c_address 값이 필요합니다")
	}
	if config.SensorStateDBPath == "" {
		return fmt.Errorf("sensor.state_db_path 값이 필요합니다")
	}
	if config.SensorHeaterTempC == 0 {
		return fmt.Errorf("sensor.heater_temp_c 값이 필요합니다")
	}
	if config.SensorHeaterDuration <= 0 {
		return fmt.Errorf("sensor.heater_duration 값이 필요합니다")
	}
	if config.SensorPollInterval <= 0 {
		return fmt.Errorf("sensor.poll_interval 값이 필요합니다")
	}
	if config.SensorStateCheckpointValidSamples <= 0 {
		return fmt.Errorf("sensor.state_checkpoint_valid_samples 값이 필요합니다")
	}
	if config.SensorAirQualityHistoryLimit <= 0 {
		return fmt.Errorf("sensor.air_quality.history_limit 값이 필요합니다")
	}
	if config.SensorAirQualityStabilizationDuration <= 0 {
		return fmt.Errorf("sensor.air_quality.stabilization_duration 값이 필요합니다")
	}
	if config.SensorAirQualityLearningDuration <= 0 {
		return fmt.Errorf("sensor.air_quality.learning_duration 값이 필요합니다")
	}
	if config.SensorAirQualityStabilizationValidSampleGoal <= 0 {
		return fmt.Errorf("sensor.air_quality.stabilization_valid_sample_goal 값이 필요합니다")
	}
	if config.SensorAirQualityLearningValidSampleGoal <= 0 {
		return fmt.Errorf("sensor.air_quality.learning_valid_sample_goal 값이 필요합니다")
	}

	return nil
}

func parseI2CAddress(value string) (uint8, error) {
	parsed, err := strconv.ParseUint(value, 0, 8)
	if err != nil {
		return 0, fmt.Errorf("sensor.i2c_address 값이 올바르지 않습니다: %w", err)
	}

	return uint8(parsed), nil
}
