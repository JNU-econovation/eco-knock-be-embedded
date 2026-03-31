package common

import (
	"context"
	"errors"
	"time"

	healthpb "google.golang.org/grpc/health/grpc_health_v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GrpcConfig struct {
	Address string
	Timeout time.Duration
}

func NewGRPCConnect(ctx context.Context, config GrpcConfig) (*grpc.ClientConn, error) {
	if config.Address == "" {
		return nil, errors.New("grpc address is required")
	}

	if config.Timeout <= 0 {
		return nil, errors.New("grpc timeout is required")
	}

	conn, err := grpc.NewClient(
		config.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	if err != nil {
		return nil, err
	}

	healthClient := healthpb.NewHealthClient(conn)

	checkCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	_, err = healthClient.Check(checkCtx, &healthpb.HealthCheckRequest{})

	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	return conn, nil
}
