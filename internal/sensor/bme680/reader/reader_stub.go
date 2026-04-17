//go:build !linux

package reader

import (
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	"eco-knock-be-embedded/internal/sensor/dto"
	"errors"
	"math"
	"sync"
	"time"
)

var (
	ErrUnsupportedPlatform = errors.New("BME680 센서 접근은 Linux 환경이 필요합니다")
	ErrSensorClosed        = errors.New("BME680 센서가 이미 종료되었습니다")
)

type Sensor struct {
	mu       sync.Mutex
	closed   bool
	readings int
	started  time.Time
}

func Open(cfg bme680config.Config) (*Sensor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &Sensor{
		started: time.Now(),
	}, nil
}

func (sensor *Sensor) Read() (dto.SampleDTO, error) {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	if sensor.closed {
		return dto.SampleDTO{}, ErrSensorClosed
	}

	sensor.readings++
	phase := float64(sensor.readings) / 4

	return dto.SampleDTO{
		TemperatureC:     24.0 + math.Sin(phase)*1.8,
		HumidityRH:       45.0 + math.Sin(phase/2)*6.5,
		GasResistanceOhm: 11800 + math.Cos(phase/3)*900,
		Status:           0xB0,
		GasValid:         true,
		HeatStable:       true,
		MeasuredAt:       time.Now(),
	}, nil
}

func (sensor *Sensor) Close() error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	sensor.closed = true
	return nil
}
