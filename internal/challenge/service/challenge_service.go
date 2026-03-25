package service

import (
	"eco-knock-be-embedded/internal/challenge/client"
	"eco-knock-be-embedded/internal/challenge/dto/client/response"
)

type ChallengeService struct {
	challengeClient *client.ChallengeClient
}

func NewChallengeService(challengeClient *client.ChallengeClient) *ChallengeService {
	return &ChallengeService{
		challengeClient: challengeClient,
	}
}

func (service *ChallengeService) RequestTicket(accessToken string) (*response.ChallengeCentralBackendResponse, error) {
	return service.challengeClient.RequestTicket(accessToken)
}
