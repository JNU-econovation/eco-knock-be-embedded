package reader

import (
	"fmt"

	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	"eco-knock-be-embedded/internal/sensor/bme680/reader/interfaces"
)

const (
	ModeReal = "real"
	ModeStub = "stub"
)

func Open(cfg bme680config.Config, mode string) (interfaces.Reader, error) {
	switch mode {
	case ModeReal:
		return openReal(cfg)
	case ModeStub:
		return openStub()
	default:
		return nil, fmt.Errorf("BME680 reader_mode 값은 real 또는 stub이어야 합니다: %s", mode)
	}
}
