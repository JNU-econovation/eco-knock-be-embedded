//go:build !linux

package test

import (
	"context"
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/service"
	"testing"
	"time"
)

func TestStatusUsesStubClientOnNonLinux(t *testing.T) {
	t.Parallel()

	conf, err := airconfig.New("127.0.0.1:54321", "00112233445566778899aabbccddeeff", time.Second)
	if err != nil {
		t.Fatalf("unexpected new config error: %v", err)
	}

	airPurifierService, err := service.New(conf)
	if err != nil {
		t.Fatalf("unexpected new service error: %v", err)
	}

	status, err := airPurifierService.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected status error: %v", err)
	}

	if !status.IsOn {
		t.Fatal("expected purifier stub to be on")
	}

	if status.AQI == 0 {
		t.Fatal("expected non-zero stub AQI")
	}

	if err := airPurifierService.SetPower(context.Background(), false); err != nil {
		t.Fatalf("unexpected set power error: %v", err)
	}

	status, err = airPurifierService.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected status error after set power: %v", err)
	}

	if status.IsOn {
		t.Fatal("expected purifier stub to be off after set power")
	}
}
