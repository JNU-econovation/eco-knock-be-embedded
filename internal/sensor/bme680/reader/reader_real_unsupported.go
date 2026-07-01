//go:build !linux

package reader

import (
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	"eco-knock-be-embedded/internal/sensor/bme680/reader/interfaces"
)

func openReal(_ bme680config.Config) (interfaces.Reader, error) {
	return nil, ErrUnsupportedPlatform
}
