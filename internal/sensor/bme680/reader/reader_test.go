package reader

import (
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	"testing"
)

func TestOpenStubReadsSampleWithoutHardwareConfig(t *testing.T) {
	t.Parallel()

	sensor, err := Open(bme680config.Config{}, ModeStub)
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
}
