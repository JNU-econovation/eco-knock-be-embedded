package client

import (
	"context"
	"errors"
)

type FakeClient struct {
	requests  [][]byte
	responses [][]byte
}

func NewFakeClient(responses ...[]byte) *FakeClient {
	client := &FakeClient{
		responses: make([][]byte, 0, len(responses)),
	}

	for _, response := range responses {
		client.responses = append(client.responses, append([]byte(nil), response...))
	}

	return client
}

func (client *FakeClient) Requests() [][]byte {
	requests := make([][]byte, 0, len(client.requests))
	for _, request := range client.requests {
		requests = append(requests, append([]byte(nil), request...))
	}

	return requests
}

func (client *FakeClient) HandShake(_ context.Context, request []byte) ([]byte, error) {
	return client.respond(request)
}

func (client *FakeClient) Send(_ context.Context, request []byte) ([]byte, error) {
	return client.respond(request)
}

func (client *FakeClient) respond(request []byte) ([]byte, error) {
	client.requests = append(client.requests, append([]byte(nil), request...))

	if len(client.responses) == 0 {
		return nil, errors.New("no response configured")
	}

	response := client.responses[0]
	client.responses = client.responses[1:]

	return append([]byte(nil), response...), nil
}
