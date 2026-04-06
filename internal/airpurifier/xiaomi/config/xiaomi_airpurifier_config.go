package config

import (
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/constant"
	"encoding/hex"
	"strings"
	"time"
)

type Config struct {
	Address string
	Token   []byte
	Timeout time.Duration
}

func New(address string, token string, timeout time.Duration) (Config, error) {
	parsedToken, err := parseToken(token)

	if err != nil {
		return Config{}, err
	}

	return Config{
		Address: address,
		Token:   parsedToken,
		Timeout: timeout,
	}, nil
}

func parseToken(value string) ([]byte, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return nil, constant.ErrTokenRequired
	}

	if len(normalized) != 32 {
		return nil, constant.ErrInvalidTokenLength
	}

	token, err := hex.DecodeString(normalized)
	if err != nil {
		return nil, err
	}

	return token, nil
}
