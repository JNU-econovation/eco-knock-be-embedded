package main

import (
	"eco-knock-be-embedded/internal/challenge/client"
	"eco-knock-be-embedded/internal/challenge/handler"
	"eco-knock-be-embedded/internal/challenge/router"
	"eco-knock-be-embedded/internal/challenge/service"
	"eco-knock-be-embedded/internal/common/config"
	"eco-knock-be-embedded/internal/common/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.Use(middleware.HandleErrors())

	conf := config.MustLoad()

	challengeClient := client.NewChallengeClient(conf.CentralBackendUrl)
	challengeService := service.NewChallengeService(challengeClient)
	challengeHandler := handler.NewChallengeHandler(challengeService)

	router.RegisterChallengeRoutes(r, challengeHandler)

	r.Run(":8080")
}
