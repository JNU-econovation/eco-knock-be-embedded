//go:build !linux

package reader

import (
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	"errors"
	"testing"
)

func TestOpenRealReturnsUnsupportedOnNonLinux(t *testing.T) {
	t.Parallel()

	_, err := Open(bme680config.Config{}, ModeReal)
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("expected unsupported platform error, got %v", err)
	}
}
