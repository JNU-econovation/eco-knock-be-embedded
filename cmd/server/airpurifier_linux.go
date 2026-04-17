//go:build linux

package main

import (
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	commonconfig "eco-knock-be-embedded/internal/common/config"
	"log"
)

func resolveAirPurifierConfig(commonConfig commonconfig.CommonConfig) (airconfig.Config, bool, error) {
	if commonConfig.AirPurifierAddress == "" || commonConfig.AirPurifierToken == "" || commonConfig.AirPurifierTimeout <= 0 {
		log.Printf("공기청정기 gRPC 서버를 건너뜁니다: 설정이 완전하지 않습니다")
		return airconfig.Config{}, false, nil
	}

	conf, err := airconfig.New(commonConfig.AirPurifierAddress, commonConfig.AirPurifierToken, commonConfig.AirPurifierTimeout)
	if err != nil {
		return airconfig.Config{}, false, err
	}

	return conf, true, nil
}
