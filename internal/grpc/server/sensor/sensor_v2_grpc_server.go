package sensor

import (
	"context"

	"eco-knock-be-embedded/internal/common/apperror"
	sensorv2pb "eco-knock-be-embedded/internal/grpc/pb/sensor/v2"
	sensorservice "eco-knock-be-embedded/internal/sensor/service"
)

type V2GRPCServer struct {
	sensorv2pb.UnimplementedSensorServiceServer
	sensorService *sensorservice.SensorService
}

func NewV2GRPCServer(service *sensorservice.SensorService) (*V2GRPCServer, error) {
	if service == nil {
		return nil, ErrSensorServiceRequired
	}

	return &V2GRPCServer{
		sensorService: service,
	}, nil
}

func (server *V2GRPCServer) GetCurrentSensor(
	_ context.Context,
	_ *sensorv2pb.GetCurrentSensorRequest,
) (*sensorv2pb.GetCurrentSensorResponse, error) {
	result := server.sensorService.Read()
	if result.Err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.SensorReadFailed, result.Err))
	}

	return &sensorv2pb.GetCurrentSensorResponse{
		TemperatureC:             result.Snapshot.Sample.TemperatureC,
		HumidityRh:               result.Snapshot.Sample.HumidityRH,
		GasResistanceOhm:         result.Snapshot.Sample.GasResistanceOhm,
		Status:                   uint32(result.Snapshot.Sample.Status),
		GasValid:                 result.Snapshot.Sample.GasValid,
		HeatStable:               result.Snapshot.Sample.HeatStable,
		MeasuredAtUnixMs:         result.Snapshot.Sample.MeasuredAt.UnixMilli(),
		StaticIaq:                result.Snapshot.AirQuality.StaticIAQ,
		EstimatedEco2Ppm:         result.Snapshot.AirQuality.EstimatedECO2PPM,
		EstimatedBvocPpm:         result.Snapshot.AirQuality.EstimatedBVOCPPM,
		Accuracy:                 result.Snapshot.AirQuality.Accuracy,
		StabilizationProgressPct: result.Snapshot.AirQuality.StabilizationProgressPct,
		GasPercentage:            result.Snapshot.AirQuality.GasPercentage,
		LearningCompleteAtUnixMs: result.Snapshot.AirQuality.LearningCompleteAt.UnixMilli(),
	}, nil
}
