package config

import (
	"testing"
	"time"
)

func TestConfigValidateAcceptsExplicitConfig(t *testing.T) {
	t.Parallel()

	err := validConfig().Validate()
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestConfigValidateRejectsMissingFields(t *testing.T) {
	t.Parallel()

	err := (Config{}).Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}
}

func TestConfigValidateRejectsInvalidOverrides(t *testing.T) {
	t.Parallel()

	config := validConfig()
	config.I2CAddress = 0x11
	config.PollInterval = 0

	err := config.Validate()
	if err == nil {
		t.Fatal("Validate() error = nil, want validation error")
	}
}

func validConfig() Config {
	return Config{
		I2CDevice:    "/dev/i2c-1",
		I2CAddress:   0x10,
		PollInterval: time.Second,
	}
}
