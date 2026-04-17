package sensor

import (
	"context"
	"errors"

	"eco-knock-be-embedded/internal/common/apperror"
	sensorv1pb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
	sensorservice "eco-knock-be-embedded/internal/sensor/service"
)

var ErrSensorServiceRequired = errors.New("센서 서비스가 필요합니다")

type V1GRPCServer struct {
	sensorv1pb.UnimplementedSensorServiceServer
	sensorService *sensorservice.SensorService
}

func NewV1GRPCServer(service *sensorservice.SensorService) (*V1GRPCServer, error) {
	if service == nil {
		return nil, ErrSensorServiceRequired
	}

	return &V1GRPCServer{
		sensorService: service,
	}, nil
}

func (server *V1GRPCServer) GetCurrentSensor(
	_ context.Context,
	_ *sensorv1pb.GetCurrentSensorRequest,
) (*sensorv1pb.GetCurrentSensorResponse, error) {
	result := server.sensorService.Read()
	if result.Err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.SensorReadFailed, result.Err))
	}

	return &sensorv1pb.GetCurrentSensorResponse{
		TemperatureC:     result.Snapshot.Sample.TemperatureC,
		HumidityRh:       result.Snapshot.Sample.HumidityRH,
		GasResistanceOhm: result.Snapshot.Sample.GasResistanceOhm,
		Status:           uint32(result.Snapshot.Sample.Status),
		GasValid:         result.Snapshot.Sample.GasValid,
		HeatStable:       result.Snapshot.Sample.HeatStable,
		MeasuredAtUnixMs: result.Snapshot.Sample.MeasuredAt.UnixMilli(),
	}, nil
}
