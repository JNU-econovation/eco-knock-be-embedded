//go:build linux

package main

import (
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	commonconfig "eco-knock-be-embedded/internal/common/config"
	"log"
)

func resolveAirPurifierConfig(commonConfig commonconfig.CommonConfig) (airconfig.Config, bool, error) {
	if commonConfig.AirPurifierAddress == "" || commonConfig.AirPurifierToken == "" || commonConfig.AirPurifierTimeout <= 0 {
		log.Printf("air purifier grpc server skipped: configuration is incomplete")
		return airconfig.Config{}, false, nil
	}

	conf, err := airconfig.New(commonConfig.AirPurifierAddress, commonConfig.AirPurifierToken, commonConfig.AirPurifierTimeout)
	if err != nil {
		return airconfig.Config{}, false, err
	}

	return conf, true, nil
}
