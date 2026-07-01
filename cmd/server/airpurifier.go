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

type airPurifierRuntimeConfig struct {
	clientMode string
	config     airconfig.Config
	enabled    bool
}

func resolveAirPurifierConfig(commonConfig commonconfig.CommonConfig) (airPurifierRuntimeConfig, error) {
	switch commonConfig.AirPurifierClientMode {
	case commonconfig.AirPurifierClientModeDisabled:
		log.Printf("공기청정기 gRPC 서버를 건너뜁니다: air_purifier.client_mode=disabled")
		return airPurifierRuntimeConfig{}, nil
	case commonconfig.AirPurifierClientModeStub:
		conf, err := airconfig.New(stubAirPurifierAddress, stubAirPurifierToken, stubAirPurifierTimeout)
		if err != nil {
			return airPurifierRuntimeConfig{}, err
		}
		log.Printf("공기청정기 gRPC 서버가 stub 클라이언트로 동작합니다")
		return airPurifierRuntimeConfig{
			clientMode: commonConfig.AirPurifierClientMode,
			config:     conf,
			enabled:    true,
		}, nil
	case commonconfig.AirPurifierClientModeReal:
		conf, err := airconfig.New(commonConfig.AirPurifierAddress, commonConfig.AirPurifierToken, commonConfig.AirPurifierTimeout)
		if err != nil {
			return airPurifierRuntimeConfig{}, err
		}
		return airPurifierRuntimeConfig{
			clientMode: commonConfig.AirPurifierClientMode,
			config:     conf,
			enabled:    true,
		}, nil
	default:
		return airPurifierRuntimeConfig{}, nil
	}
}
