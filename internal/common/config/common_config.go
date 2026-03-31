package config

import (
	"log"
	"os"
	"strconv"
)

type CommonConfig struct {
	CentralBackendURL          string
	ServerPort                 int
	AllowCentralBackendFailure bool
}

func MustLoad() CommonConfig {
	allowCentralBackendFailure := parseRequiredBool("ALLOW_CENTRAL_BACKEND_FAILURE", false)

	centralBackendURL := os.Getenv("CENTRAL_BACKEND_BASE_URL")
	if centralBackendURL == "" && !allowCentralBackendFailure {
		log.Fatal("CENTRAL_BACKEND_BASE_URL 환경변수가 필요합니다.")
	}

	serverPortString := os.Getenv("SERVER_PORT")
	if serverPortString == "" {
		log.Fatal("SERVER_PORT 환경변수가 필요합니다.")
	}

	serverPort, err := strconv.Atoi(serverPortString)
	if err != nil {
		log.Fatal("잘못된 SERVER_PORT 값 : ", err)
	}

	return CommonConfig{
		CentralBackendURL:          centralBackendURL,
		ServerPort:                 serverPort,
		AllowCentralBackendFailure: allowCentralBackendFailure,
	}
}

func parseRequiredBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		log.Fatal("잘못된 ", key, " 값 : ", err)
	}

	return parsed
}
