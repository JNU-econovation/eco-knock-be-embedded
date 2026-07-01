package config

import (
	"strings"
	"testing"
	"time"
)

func TestValidateAllowsStubModesWithoutHardwareConfig(t *testing.T) {
	t.Parallel()

	config := validCommonConfig()
	config.SensorReaderMode = ReaderModeStub
	config.SensorI2CDevice = ""
	config.SensorI2CAddress = 0
	config.LightSensorReaderMode = ReaderModeStub
	config.LightSensorI2CDevice = ""
	config.LightSensorI2CAddress = 0
	config.AirPurifierClientMode = AirPurifierClientModeStub
	config.AirPurifierAddress = ""
	config.AirPurifierToken = ""
	config.AirPurifierTimeout = 0

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}

func TestValidateRequiresHardwareConfigForRealModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		mutate    func(*CommonConfig)
		wantError string
	}{
		{
			name: "sensor i2c device",
			mutate: func(config *CommonConfig) {
				config.SensorI2CDevice = ""
			},
			wantError: "sensor.i2c_device",
		},
		{
			name: "light sensor i2c address",
			mutate: func(config *CommonConfig) {
				config.LightSensorI2CAddress = 0
			},
			wantError: "light_sensor.i2c_address",
		},
		{
			name: "air purifier token",
			mutate: func(config *CommonConfig) {
				config.AirPurifierToken = ""
			},
			wantError: "air_purifier.token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := validCommonConfig()
			tt.mutate(&config)

			err := config.Validate()
			if err == nil {
				t.Fatal("expected validate error")
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error to contain %q, got %v", tt.wantError, err)
			}
		})
	}
}

func TestValidateRejectsInvalidModes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		mutate func(*CommonConfig)
	}{
		{
			name: "sensor",
			mutate: func(config *CommonConfig) {
				config.SensorReaderMode = "auto"
			},
		},
		{
			name: "light sensor",
			mutate: func(config *CommonConfig) {
				config.LightSensorReaderMode = "auto"
			},
		},
		{
			name: "air purifier",
			mutate: func(config *CommonConfig) {
				config.AirPurifierClientMode = "auto"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := validCommonConfig()
			tt.mutate(&config)

			if err := config.Validate(); err == nil {
				t.Fatal("expected validate error")
			}
		})
	}
}

func TestValidateAllowsDisabledAirPurifierWithoutDeviceConfig(t *testing.T) {
	t.Parallel()

	config := validCommonConfig()
	config.AirPurifierClientMode = AirPurifierClientModeDisabled
	config.AirPurifierAddress = ""
	config.AirPurifierToken = ""
	config.AirPurifierTimeout = 0

	if err := config.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
}

func validCommonConfig() CommonConfig {
	return CommonConfig{
		ServerHTTPPort:                               19090,
		ServerGRPCPort:                               6565,
		SensorReaderMode:                             ReaderModeReal,
		SensorI2CDevice:                              "/dev/i2c-1",
		SensorI2CAddress:                             0x76,
		SensorStateDBPath:                            "data/sensor_state.db",
		SensorHeaterTempC:                            300,
		SensorHeaterDuration:                         100 * time.Millisecond,
		SensorPollInterval:                           time.Second,
		SensorStateCheckpointValidSamples:            60,
		SensorAirQualityHistoryLimit:                 3600,
		SensorAirQualityStabilizationDuration:        5 * time.Minute,
		SensorAirQualityLearningDuration:             time.Hour,
		SensorAirQualityStabilizationValidSampleGoal: 300,
		SensorAirQualityLearningValidSampleGoal:      3600,
		LightSensorReaderMode:                        ReaderModeReal,
		LightSensorI2CDevice:                         "/dev/i2c-1",
		LightSensorI2CAddress:                        0x10,
		LightSensorPollInterval:                      time.Second,
		AirPurifierClientMode:                        AirPurifierClientModeReal,
		AirPurifierAddress:                           "192.168.0.50:54321",
		AirPurifierToken:                             "0123456789abcdef0123456789abcdef",
		AirPurifierTimeout:                           3 * time.Second,
	}
}
