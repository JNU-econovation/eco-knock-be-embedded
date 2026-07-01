package reader

import (
	"fmt"

	veml7700config "eco-knock-be-embedded/internal/lightsensor/veml7700/config"
	"eco-knock-be-embedded/internal/lightsensor/veml7700/reader/interfaces"
)

const (
	ModeReal = "real"
	ModeStub = "stub"
)

func Open(cfg veml7700config.Config, mode string) (interfaces.Reader, error) {
	switch mode {
	case ModeReal:
		return openReal(cfg)
	case ModeStub:
		return openStub()
	default:
		return nil, fmt.Errorf("VEML7700 reader_mode 값은 real 또는 stub이어야 합니다: %s", mode)
	}
}
