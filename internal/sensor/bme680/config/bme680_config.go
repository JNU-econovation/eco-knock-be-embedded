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
		return invalidConfigError("i2c device가 설정되지 않았습니다")
	case config.I2CAddress != 0x76 && config.I2CAddress != 0x77:
		return invalidConfigError("i2c address는 0x76 또는 0x77이어야 합니다")
	case config.HeaterTempC == 0:
		return invalidConfigError("heater temperature는 0보다 커야 합니다")
	case config.HeaterTempC > maxHeaterTempC:
		return invalidConfigError("heater temperature는 400C 이하여야 합니다")
	case config.HeaterDuration < minHeaterDuration:
		return invalidConfigError("heater duration은 최소 1ms 이상이어야 합니다")
	case config.HeaterDuration > maxHeaterDuration:
		return invalidConfigError("heater duration은 65535ms 이하여야 합니다")
	}

	return nil
}

type invalidConfigError string

func (err invalidConfigError) Error() string {
	return string(err)
}
