package main

import (
	commonconfig "eco-knock-be-embedded/internal/common/config"
	"testing"
)

func TestResolveAirPurifierConfigDisabled(t *testing.T) {
	t.Parallel()

	runtimeConfig, err := resolveAirPurifierConfig(commonconfig.CommonConfig{
		AirPurifierClientMode: commonconfig.AirPurifierClientModeDisabled,
	})
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if runtimeConfig.enabled {
		t.Fatal("expected air purifier to be disabled")
	}
}

func TestResolveAirPurifierConfigStubWithoutDeviceConfig(t *testing.T) {
	t.Parallel()

	runtimeConfig, err := resolveAirPurifierConfig(commonconfig.CommonConfig{
		AirPurifierClientMode: commonconfig.AirPurifierClientModeStub,
	})
	if err != nil {
		t.Fatalf("unexpected resolve error: %v", err)
	}
	if !runtimeConfig.enabled {
		t.Fatal("expected air purifier to be enabled")
	}
	if runtimeConfig.clientMode != commonconfig.AirPurifierClientModeStub {
		t.Fatalf("expected stub mode, got %q", runtimeConfig.clientMode)
	}
}
