//go:build linux

package client

import (
	"context"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/util"
)

type realClient struct {
	config config.Config
}

func newRealClient(config config.Config) (Requester, error) {
	return &realClient{
		config: config,
	}, nil
}

func (client *realClient) HandShake(
	ctx context.Context,
	helloPacket []byte,
) ([]byte, error) {
	return util.RequestReply(ctx, client.config.Address, helloPacket, client.config.Timeout)
}

func (client *realClient) Send(
	ctx context.Context,
	request []byte,
) ([]byte, error) {
	return util.RequestReply(ctx, client.config.Address, request, client.config.Timeout)
}
