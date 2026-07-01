//go:build !linux

package client

import (
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	"errors"
	"testing"
	"time"
)

func TestNewRealReturnsUnsupportedOnNonLinux(t *testing.T) {
	t.Parallel()

	conf, err := airconfig.New("127.0.0.1:54321", "00112233445566778899aabbccddeeff", time.Second)
	if err != nil {
		t.Fatalf("unexpected config error: %v", err)
	}

	_, err = New(conf, ModeReal)
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("expected unsupported platform error, got %v", err)
	}
}
