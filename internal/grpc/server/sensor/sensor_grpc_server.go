package sensor

import (
	"context"
	"eco-knock-be-embedded/internal/common/apperror"
	"eco-knock-be-embedded/internal/sensor/service"
	"errors"

	sensorpb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
)

var ErrStreamingServiceRequired = errors.New("sensor streaming service is required")

type GRPCServer struct {
	sensorpb.UnimplementedSensorServiceServer
	sensorService *service.SensorService
}

func NewGRPCServer(service *service.SensorService) (*GRPCServer, error) {
	if service == nil {
		return nil, ErrStreamingServiceRequired
	}

	return &GRPCServer{
		sensorService: service,
	}, nil
}

func (server *GRPCServer) GetCurrentSensor(
	_ context.Context,
	_ *sensorpb.GetCurrentSensorRequest,
) (*sensorpb.GetCurrentSensorResponse, error) {
	result := server.sensorService.Read()
	if result.Err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.SensorReadFailed, result.Err))
	}

	return &sensorpb.GetCurrentSensorResponse{
		TemperatureC:     result.Sample.TemperatureC,
		HumidityRh:       result.Sample.HumidityRH,
		GasResistanceOhm: result.Sample.GasResistanceOhm,
		Status:           uint32(result.Sample.Status),
		GasValid:         result.Sample.GasValid,
		HeatStable:       result.Sample.HeatStable,
		MeasuredAtUnixMs: result.Sample.MeasuredAt.UnixMilli(),
	}, nil
}
