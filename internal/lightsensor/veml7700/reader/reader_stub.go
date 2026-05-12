//go:build !linux || lightsensor_stub

package reader

import (
	"eco-knock-be-embedded/internal/lightsensor/dto"
	veml7700config "eco-knock-be-embedded/internal/lightsensor/veml7700/config"
	"errors"
	"math"
	"sync"
	"time"
)

var (
	ErrUnsupportedPlatform = errors.New("VEML7700 조도 센서 접근은 Linux 환경이 필요합니다")
	ErrSensorClosed        = errors.New("VEML7700 조도 센서가 이미 종료되었습니다")
)

type Sensor struct {
	mu       sync.Mutex
	closed   bool
	readings int
}

func Open(cfg veml7700config.Config) (*Sensor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Sensor{}, nil
}

func (sensor *Sensor) Read() (dto.SampleDTO, error) {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	if sensor.closed {
		return dto.SampleDTO{}, ErrSensorClosed
	}

	sensor.readings++
	phase := float64(sensor.readings) / 5
	lux := 180 + math.Sin(phase)*45

	return dto.SampleDTO{
		Lux:        lux,
		RawALS:     uint16(lux / alsResolution),
		RawWhite:   uint16((lux / alsResolution) * 1.08),
		MeasuredAt: time.Now(),
	}, nil
}

func (sensor *Sensor) Close() error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	sensor.closed = true
	return nil
}
