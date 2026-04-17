//go:build !linux

package main

import (
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	commonconfig "eco-knock-be-embedded/internal/common/config"
	"log"
	"time"
)

const (
	stubAirPurifierAddress = "127.0.0.1:54321"
	stubAirPurifierToken   = "00112233445566778899aabbccddeeff"
)

var stubAirPurifierTimeout = 3 * time.Second

func resolveAirPurifierConfig(commonConfig commonconfig.CommonConfig) (airconfig.Config, bool, error) {
	address := commonConfig.AirPurifierAddress
	if address == "" {
		address = stubAirPurifierAddress
	}

	token := commonConfig.AirPurifierToken
	if token == "" {
		token = stubAirPurifierToken
	}

	timeout := commonConfig.AirPurifierTimeout
	if timeout <= 0 {
		timeout = stubAirPurifierTimeout
	}

	log.Printf("air purifier grpc server uses stub client on non-linux platform")

	conf, err := airconfig.New(address, token, timeout)
	if err != nil {
		return airconfig.Config{}, false, err
	}

	return conf, true, nil
}
