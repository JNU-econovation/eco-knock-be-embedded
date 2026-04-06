package util

import (
	"context"
	"net"
	"time"
)

func RequestReply(
	ctx context.Context,
	address string,
	request []byte,
	timeout time.Duration,
) ([]byte, error) {
	dialer := net.Dialer{}
	conn, err := dialer.DialContext(ctx, "udp", address)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	if err := conn.SetDeadline(resolveDeadline(ctx, timeout)); err != nil {
		return nil, err
	}

	if _, err := conn.Write(request); err != nil {
		return nil, err
	}

	response := make([]byte, 4096)
	n, err := conn.Read(response)
	if err != nil {
		return nil, err
	}

	return append([]byte(nil), response[:n]...), nil
}

func resolveDeadline(ctx context.Context, timeout time.Duration) time.Time {
	deadline := time.Now().Add(timeout)
	if ctxDeadline, ok := ctx.Deadline(); ok && ctxDeadline.Before(deadline) {
		return ctxDeadline
	}

	return deadline
}
