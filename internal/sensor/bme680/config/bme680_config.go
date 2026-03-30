package config

import "time"

const (
	minHeaterDuration = time.Millisecond
	maxHeaterDuration = 0xFFFF * time.Millisecond
	maxHeaterTempC    = 400
)

// Config Config는 BME680과 통신하기 위해 필요한 하드웨어 설정을 담는다.
type Config struct {
	I2CDevice      string
	I2CAddress     uint8
	HeaterTempC    uint16
	HeaterDuration time.Duration
	AmbientTempC   int8
}

func (config Config) Validate() error {
	switch {
	case config.I2CDevice == "":
		return invalidConfigError("i2c device not configured")
	case config.I2CAddress != 0x76 && config.I2CAddress != 0x77:
		return invalidConfigError("i2c address must be 0x76 or 0x77")
	case config.HeaterTempC == 0:
		return invalidConfigError("heater temperature must be greater than zero")
	case config.HeaterTempC > maxHeaterTempC:
		return invalidConfigError("heater temperature must be <= 400C")
	case config.HeaterDuration < minHeaterDuration:
		return invalidConfigError("heater duration must be at least 1ms")
	case config.HeaterDuration > maxHeaterDuration:
		return invalidConfigError("heater duration must be <= 65535ms")
	}

	return nil
}

type invalidConfigError string

func (err invalidConfigError) Error() string {
	return string(err)
}
