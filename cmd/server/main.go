package main

import (
	"eco-knock-be-embedded/internal/common/config"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const configPath = "application.yaml"

func main() {
	envFile := os.Getenv("APP_ENV_FILE")
	if envFile == "" {
		envFile = ".env"
	}
	_ = godotenv.Load(envFile)

	if err := run(); err != nil {
		log.Fatalf("서버 실행에 실패했습니다: %v", err)
	}
}

func run() error {
	conf := config.MustLoad(configPath)

	stopGRPCServer, err := startGRPCServer(conf)
	if err != nil {
		return err
	}
	defer stopGRPCServer()

	return startRESTServer(conf)
}
