package config

import (
	"log"
	"os"
)

type CommonConfig struct {
	CentralBackendUrl string
}

func MustLoad() CommonConfig {
	baseURL := os.Getenv("CENTRAL_BACKEND_BASE_URL")
	if baseURL == "" {
		log.Fatal("CENTRAL_BACKEND_BASE_URL 환경변수가 필요합니다.")
	}

	return CommonConfig{
		CentralBackendUrl: baseURL,
	}
}
