package reader

import (
	veml7700config "eco-knock-be-embedded/internal/lightsensor/veml7700/config"
	"testing"
)

func TestOpenStubReadsSampleWithoutHardwareConfig(t *testing.T) {
	t.Parallel()

	sensor, err := Open(veml7700config.Config{}, ModeStub)
	if err != nil {
		t.Fatalf("unexpected open stub error: %v", err)
	}
	defer func() {
		_ = sensor.Close()
	}()

	sample, err := sensor.Read()
	if err != nil {
		t.Fatalf("unexpected read error: %v", err)
	}
	if sample.MeasuredAt.IsZero() {
		t.Fatal("expected measured time")
	}
	if sample.Lux == 0 {
		t.Fatal("expected non-zero lux")
	}
}
