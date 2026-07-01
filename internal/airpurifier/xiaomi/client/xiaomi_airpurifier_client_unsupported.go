//go:build !linux

package client

import "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"

func newRealClient(_ config.Config) (Requester, error) {
	return nil, ErrUnsupportedPlatform
}
