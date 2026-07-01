package client

import (
	"context"
	"errors"

	"eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
)

const (
	ModeReal = "real"
	ModeStub = "stub"
)

var (
	ErrUnsupportedPlatform = errors.New("샤오미 공기청정기 실제 클라이언트는 Linux 환경이 필요합니다")
	ErrInvalidClientMode   = errors.New("공기청정기 client_mode 값은 real 또는 stub이어야 합니다")
)

type Requester interface {
	HandShake(ctx context.Context, helloPacket []byte) ([]byte, error)
	Send(ctx context.Context, request []byte) ([]byte, error)
}

func New(config config.Config, mode string) (Requester, error) {
	switch mode {
	case ModeReal:
		return newRealClient(config)
	case ModeStub:
		return newStubClient(config), nil
	default:
		return nil, ErrInvalidClientMode
	}
}
