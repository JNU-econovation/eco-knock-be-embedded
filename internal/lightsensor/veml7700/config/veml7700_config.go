package config

import "time"

const defaultAddress = 0x10

// Config Config는 VEML7700과 통신하기 위해 필요한 하드웨어 설정을 담는다.
type Config struct {
	I2CDevice    string
	I2CAddress   uint8
	PollInterval time.Duration
}

func (config Config) Validate() error {
	switch {
	case config.I2CDevice == "":
		return invalidConfigError("i2c device가 설정되지 않았습니다")
	case config.I2CAddress != defaultAddress:
		return invalidConfigError("i2c address는 0x10이어야 합니다")
	case config.PollInterval <= 0:
		return invalidConfigError("poll interval은 0보다 커야 합니다")
	}

	return nil
}

type invalidConfigError string

func (err invalidConfigError) Error() string {
	return string(err)
}
