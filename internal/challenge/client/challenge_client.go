package client

import (
	"eco-knock-be-embedded/internal/challenge/dto/client/response"
	"eco-knock-be-embedded/internal/common/apperror"
	"eco-knock-be-embedded/internal/common/constant"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type ChallengeClient struct {
	httpClient     *http.Client
	centralBaseURL string
	challengePath  string
}

func NewChallengeClient(centralBaseURL string) *ChallengeClient {
	return &ChallengeClient{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		centralBaseURL: strings.TrimRight(centralBaseURL, "/"),
		challengePath:  "/challenge",
	}
}

func (client *ChallengeClient) RequestTicket(accessToken string) (*response.ChallengeCentralBackendResponse, error) {
	if client.centralBaseURL == "" {
		return nil, apperror.New(apperror.InternalServer, fmt.Errorf("CENTRAL_BACKEND_BASE_URL is not set"))
	}

	req, err := http.NewRequest(http.MethodGet, client.centralBaseURL+client.challengePath, nil)
	if err != nil {
		return nil, apperror.New(apperror.InternalServer, err)
	}

	req.Header.Set(constant.Authorization, accessToken)

	res, err := client.httpClient.Do(req)
	if err != nil {
		return nil, apperror.New(apperror.CentralBackendUnavailable, err)
	}

	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return nil, apperror.New(apperror.Unauthorized, nil)
	}

	if res.StatusCode != http.StatusOK {
		return nil, apperror.New(
			apperror.CentralBackendUnavailable,
			fmt.Errorf("central backend returned status %d", res.StatusCode),
		)
	}

	var challengeResponse response.ChallengeCentralBackendResponse

	if err := json.NewDecoder(res.Body).Decode(&challengeResponse); err != nil {
		return nil, apperror.New(apperror.CentralBackendUnavailable, err)
	}

	return &challengeResponse, nil
}
