//go:build linux

package client

import (
	"context"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/util"
)

type Client struct {
	config config.Config
}

func New(config config.Config) *Client {
	return &Client{
		config: config,
	}
}

func (client *Client) HandShake(
	ctx context.Context,
	helloPacket []byte,
) ([]byte, error) {
	return util.RequestReply(ctx, client.config.Address, helloPacket, client.config.Timeout)
}

func (client *Client) Send(
	ctx context.Context,
	request []byte,
) ([]byte, error) {
	return util.RequestReply(ctx, client.config.Address, request, client.config.Timeout)
}
