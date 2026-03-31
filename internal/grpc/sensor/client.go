package sensor

import (
	"context"
	"errors"

	sensorpb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
	"eco-knock-be-embedded/internal/sensor/dto"
)

var ErrClientRequired = errors.New("sensor grpc client is required")

type Client struct {
	grpcClient sensorpb.SensorServiceClient
}

func NewGRPCClient(grpcClient sensorpb.SensorServiceClient) (*Client, error) {
	if grpcClient == nil {
		return nil, ErrClientRequired
	}

	return &Client{
		grpcClient: grpcClient,
	}, nil
}

func (client *Client) Report(
	ctx context.Context,
	payload dto.SensorClientRequest,
) (*sensorpb.ReportSensorResponse, error) {
	request := &sensorpb.ReportSensorRequest{
		DeviceId:         int32(payload.DeviceId),
		TemperatureC:     payload.Sample.TemperatureC,
		HumidityRh:       payload.Sample.HumidityRH,
		GasResistanceOhm: payload.Sample.GasResistanceOhm,
		MeasuredAtUnixMs: payload.Sample.MeasuredAt.UnixMilli(),
	}

	return client.grpcClient.ReportSensor(ctx, request)
}
