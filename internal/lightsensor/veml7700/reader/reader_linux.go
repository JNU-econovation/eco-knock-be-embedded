//go:build linux && !lightsensor_stub

package reader

import (
	"eco-knock-be-embedded/internal/lightsensor/dto"
	veml7700config "eco-knock-be-embedded/internal/lightsensor/veml7700/config"
	"encoding/binary"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"periph.io/x/conn/v3/i2c"
	"periph.io/x/host/v3/sysfs"
)

const (
	regALSConfig = 0x00
	regALSData   = 0x04
	regWhiteData = 0x05
	regID        = 0x07

	alsGain1x             = 0x0000
	alsIntegrationTime100 = 0x0000
	alsPowerOn            = 0x0000
	chipID                = 0xC481
	startupDelay          = 130 * time.Millisecond
)

var ErrSensorClosed = errors.New("VEML7700 조도 센서가 이미 종료되었습니다")

type Sensor struct {
	mu     sync.Mutex
	bus    i2c.BusCloser
	dev    *i2c.Dev
	config veml7700config.Config
	closed bool
}

func Open(cfg veml7700config.Config) (*Sensor, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	busNumber, err := parseBusNumber(cfg.I2CDevice)
	if err != nil {
		return nil, err
	}

	bus, err := sysfs.NewI2C(busNumber)
	if err != nil {
		return nil, fmt.Errorf("I2C 버스 %d 열기에 실패했습니다: %w", busNumber, err)
	}

	sensor := &Sensor{
		bus:    bus,
		dev:    &i2c.Dev{Bus: bus, Addr: uint16(cfg.I2CAddress)},
		config: cfg,
	}

	if err := sensor.init(); err != nil {
		_ = bus.Close()
		return nil, err
	}

	return sensor, nil
}

func (sensor *Sensor) Read() (dto.SampleDTO, error) {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	if sensor.closed {
		return dto.SampleDTO{}, ErrSensorClosed
	}

	rawALS, err := sensor.readWord(regALSData)
	if err != nil {
		return dto.SampleDTO{}, fmt.Errorf("ALS 데이터 읽기에 실패했습니다: %w", err)
	}

	rawWhite, err := sensor.readWord(regWhiteData)
	if err != nil {
		return dto.SampleDTO{}, fmt.Errorf("white 데이터 읽기에 실패했습니다: %w", err)
	}

	return dto.SampleDTO{
		Lux:        calculateLux(rawALS),
		RawALS:     rawALS,
		RawWhite:   rawWhite,
		MeasuredAt: time.Now(),
	}, nil
}

func (sensor *Sensor) Close() error {
	sensor.mu.Lock()
	defer sensor.mu.Unlock()

	if sensor.closed {
		return nil
	}

	sensor.closed = true
	return sensor.bus.Close()
}

func (sensor *Sensor) init() error {
	id, err := sensor.readWord(regID)
	if err != nil {
		return fmt.Errorf("디바이스 ID 읽기에 실패했습니다: %w", err)
	}
	if id != chipID {
		return fmt.Errorf("예상하지 못한 디바이스 ID입니다: 0x%04x", id)
	}

	config := uint16(alsGain1x | alsIntegrationTime100 | alsPowerOn)
	if err := sensor.writeWord(regALSConfig, config); err != nil {
		return fmt.Errorf("ALS 설정 쓰기에 실패했습니다: %w", err)
	}

	time.Sleep(startupDelay)
	return nil
}

func (sensor *Sensor) readWord(reg byte) (uint16, error) {
	var data [2]byte
	if err := sensor.dev.Tx([]byte{reg}, data[:]); err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint16(data[:]), nil
}

func (sensor *Sensor) writeWord(reg byte, value uint16) error {
	var data [3]byte
	data[0] = reg
	binary.LittleEndian.PutUint16(data[1:], value)
	return sensor.dev.Tx(data[:], nil)
}

func parseBusNumber(ref string) (int, error) {
	base := strings.TrimSpace(ref)
	if base == "" {
		return 0, errors.New("I2C device가 설정되지 않았습니다")
	}

	base = strings.ToLower(filepath.Base(base))
	switch {
	case strings.HasPrefix(base, "i2c-"):
		base = strings.TrimPrefix(base, "i2c-")
	case strings.HasPrefix(base, "i2c"):
		base = strings.TrimPrefix(base, "i2c")
	}

	number, err := strconv.Atoi(base)
	if err != nil {
		return 0, fmt.Errorf("I2C 버스 참조 값이 올바르지 않습니다: %q", ref)
	}

	return number, nil
}
