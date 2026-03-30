package config

import (
	"log"
	"os"
	"strconv"
)

type CommonConfig struct {
	CentralBackendUrl string
	ServerPort        int
}

func MustLoad() CommonConfig {
	baseURL := os.Getenv("CENTRAL_BACKEND_BASE_URL")
	if baseURL == "" {
		log.Fatal("CENTRAL_BACKEND_BASE_URL 환경변수가 필요합니다.")
	}

	serverPortString := os.Getenv("SERVER_PORT")
	if serverPortString == "" {
		log.Fatal("SERVER_PORT 환경변수가 필요합니다.")
	}

	serverPort, err := strconv.Atoi(serverPortString)

	if err != nil {
		log.Fatal("잘못된 port : ", err)
	}

	return CommonConfig{
		CentralBackendUrl: baseURL,
		ServerPort:        serverPort,
	}
}
