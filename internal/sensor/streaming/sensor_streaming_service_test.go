package streaming

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"eco-knock-be-embedded/internal/sensor/dto"
)

func TestServiceStreamsSamples(t *testing.T) {
	t.Parallel()

	reader := &fakeReader{
		samples: []dto.SampleDTO{
			{TemperatureC: 22.4, HumidityRH: 45.1, GasResistanceOhm: 1200},
			{TemperatureC: 22.6, HumidityRH: 45.3, GasResistanceOhm: 1220},
		},
	}

	service := New(reader, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := service.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	first := waitForResult(t, stream)
	if first.Err != nil {
		t.Fatalf("first.Err = %v, want nil", first.Err)
	}
	if first.Sample.TemperatureC != 22.4 {
		t.Fatalf("first.Sample.TemperatureC = %v, want 22.4", first.Sample.TemperatureC)
	}

	second := waitForResult(t, stream)
	if second.Err != nil {
		t.Fatalf("second.Err = %v, want nil", second.Err)
	}
	if second.Sample.TemperatureC != 22.6 {
		t.Fatalf("second.Sample.TemperatureC = %v, want 22.6", second.Sample.TemperatureC)
	}
}

func TestServiceStreamsErrors(t *testing.T) {
	t.Parallel()

	readErr := errors.New("sensor read failed")
	reader := &fakeReader{
		errAt: map[int]error{
			0: readErr,
		},
	}

	service := New(reader, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stream, err := service.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	result := waitForResult(t, stream)
	if !errors.Is(result.Err, readErr) {
		t.Fatalf("result.Err = %v, want %v", result.Err, readErr)
	}
}

func TestServiceStreamRejectsNilReader(t *testing.T) {
	t.Parallel()

	service := New(nil, 10*time.Millisecond)

	stream, err := service.Stream(context.Background())
	if !errors.Is(err, ErrReaderRequired) {
		t.Fatalf("Stream() error = %v, want %v", err, ErrReaderRequired)
	}
	if stream != nil {
		t.Fatal("stream != nil, want nil")
	}
}

func TestServiceStreamClosesWhenContextCancelled(t *testing.T) {
	t.Parallel()

	service := New(&fakeReader{}, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())

	stream, err := service.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream() error = %v", err)
	}

	_ = waitForResult(t, stream)
	cancel()

	select {
	case _, ok := <-stream:
		if ok {
			t.Fatal("stream still open after cancel")
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for stream to close")
	}
}

type fakeReader struct {
	samples []dto.SampleDTO
	errAt   map[int]error
	index   int
	mu      sync.Mutex
}

func (reader *fakeReader) Read() (dto.SampleDTO, error) {
	reader.mu.Lock()
	defer reader.mu.Unlock()

	if err := reader.errAt[reader.index]; err != nil {
		reader.index++
		return dto.SampleDTO{}, err
	}

	if len(reader.samples) == 0 {
		reader.index++
		return dto.SampleDTO{}, nil
	}

	if reader.index >= len(reader.samples) {
		sample := reader.samples[len(reader.samples)-1]
		reader.index++
		return sample, nil
	}

	sample := reader.samples[reader.index]
	reader.index++
	return sample, nil
}

func (reader *fakeReader) Close() error {
	return nil
}

func waitForResult(t *testing.T, stream <-chan dto.SensorInternalDTO) dto.SensorInternalDTO {
	t.Helper()

	select {
	case result, ok := <-stream:
		if !ok {
			t.Fatal("stream closed before a result was published")
		}
		return result
	case <-time.After(300 * time.Millisecond):
		t.Fatal("timed out waiting for result")
	}

	return dto.SensorInternalDTO{}
}
