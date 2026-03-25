package handler

import (
	"eco-knock-be-embedded/internal/challenge/service"
	"eco-knock-be-embedded/internal/common/apperror"
	"eco-knock-be-embedded/internal/common/constant"

	"github.com/gin-gonic/gin"
)

type ChallengeHandler struct {
	challengeService *service.ChallengeService
}

func NewChallengeHandler(challengeService *service.ChallengeService) *ChallengeHandler {
	return &ChallengeHandler{challengeService: challengeService}
}

func (handler *ChallengeHandler) GetChallenge(context *gin.Context) error {
	accessToken := context.GetHeader(constant.Authorization)
	if accessToken == "" {
		return apperror.New(apperror.AuthorizationHeaderRequired, nil)
	}

	res, err := handler.challengeService.RequestTicket(accessToken)
	if err != nil {
		return err
	}

	context.JSON(200, res)
	return nil
}
