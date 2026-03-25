package response

type ChallengeCentralBackendResponse struct {
	Ticket     string `json:"ticket"`
	Type       string `json:"type"`
	IssuedAt   string `json:"issuedAt"`
	ExpiresAt  string `json:"expiresAt"`
	TTLSeconds int    `json:"ttlSeconds"`
	RequestID  string `json:"requestId"`
}
