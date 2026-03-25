package router

import (
	"eco-knock-be-embedded/internal/challenge/handler"
	"eco-knock-be-embedded/internal/common/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterChallengeRoutes(r *gin.Engine, challengeHandler *handler.ChallengeHandler) {
	challengeGroup := r.Group("/challenge")
	challengeGroup.GET("", middleware.WrapErrorHandler(challengeHandler.GetChallenge))
}
