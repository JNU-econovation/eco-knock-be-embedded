package client

import (
	"context"
)

type Requester interface {
	HandShake(ctx context.Context, helloPacket []byte) ([]byte, error)
	Send(ctx context.Context, request []byte) ([]byte, error)
}
