//go:build !linux

package reader

import (
	veml7700config "eco-knock-be-embedded/internal/lightsensor/veml7700/config"
	"errors"
	"testing"
)

func TestOpenRealReturnsUnsupportedOnNonLinux(t *testing.T) {
	t.Parallel()

	_, err := Open(veml7700config.Config{}, ModeReal)
	if !errors.Is(err, ErrUnsupportedPlatform) {
		t.Fatalf("expected unsupported platform error, got %v", err)
	}
}
