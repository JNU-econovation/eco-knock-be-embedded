//go:build !linux

package reader

import (
	veml7700config "eco-knock-be-embedded/internal/lightsensor/veml7700/config"
	"eco-knock-be-embedded/internal/lightsensor/veml7700/reader/interfaces"
)

func openReal(_ veml7700config.Config) (interfaces.Reader, error) {
	return nil, ErrUnsupportedPlatform
}
