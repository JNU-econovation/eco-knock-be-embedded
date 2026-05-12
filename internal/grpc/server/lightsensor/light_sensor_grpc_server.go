package lightsensor

import (
	"context"
	"errors"

	"eco-knock-be-embedded/internal/common/apperror"
	lightsensorpb "eco-knock-be-embedded/internal/grpc/pb/lightsensor"
	lightsensorservice "eco-knock-be-embedded/internal/lightsensor/service"
)

var ErrLightSensorServiceRequired = errors.New("조도 센서 서비스가 필요합니다")

type GRPCServer struct {
	lightsensorpb.UnimplementedLightSensorServiceServer
	lightSensorService *lightsensorservice.LightSensorService
}

func NewGRPCServer(service *lightsensorservice.LightSensorService) (*GRPCServer, error) {
	if service == nil {
		return nil, ErrLightSensorServiceRequired
	}

	return &GRPCServer{
		lightSensorService: service,
	}, nil
}

func (server *GRPCServer) GetCurrentLightSensor(
	_ context.Context,
	_ *lightsensorpb.GetCurrentLightSensorRequest,
) (*lightsensorpb.GetCurrentLightSensorResponse, error) {
	result := server.lightSensorService.Read()
	if result.Err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.LightSensorReadFailed, result.Err))
	}

	return &lightsensorpb.GetCurrentLightSensorResponse{
		Lux:              result.Sample.Lux,
		RawAls:           uint32(result.Sample.RawALS),
		RawWhite:         uint32(result.Sample.RawWhite),
		MeasuredAtUnixMs: result.Sample.MeasuredAt.UnixMilli(),
	}, nil
}
