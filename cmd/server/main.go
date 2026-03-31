package main

import (
	"eco-knock-be-embedded/internal/common/config"
	"eco-knock-be-embedded/internal/common/middleware"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load(".env")

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	r := gin.Default()
	r.Use(middleware.HandleErrors())

	conf := config.MustLoad()

	stopSensorReporter, err := startSensorReporter(conf)
	if err != nil {
		return err
	}
	defer stopSensorReporter()

	return r.Run(fmt.Sprintf(":%d", conf.ServerPort))
}
