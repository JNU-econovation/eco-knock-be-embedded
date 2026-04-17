package airpurifier

import (
	"context"
	"errors"

	airservice "eco-knock-be-embedded/internal/airpurifier/xiaomi/service"
	"eco-knock-be-embedded/internal/common/apperror"
	airpurifierpb "eco-knock-be-embedded/internal/grpc/pb/airpurifier/v1"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

var ErrAirPurifierServiceRequired = errors.New("공기청정기 서비스가 필요합니다")

type GRPCServer struct {
	airpurifierpb.UnimplementedAirPurifierServiceServer
	airPurifierService *airservice.XiaomiAirPurifierService
}

func NewGRPCServer(service *airservice.XiaomiAirPurifierService) (*GRPCServer, error) {
	if service == nil {
		return nil, ErrAirPurifierServiceRequired
	}

	return &GRPCServer{
		airPurifierService: service,
	}, nil
}

func (server *GRPCServer) GetCurrentAirPurifier(
	ctx context.Context,
	_ *airpurifierpb.GetCurrentAirPurifierRequest,
) (*airpurifierpb.GetCurrentAirPurifierResponse, error) {
	result, err := server.airPurifierService.Status(ctx)
	if err != nil {
		return nil, apperror.ToGRPCError(apperror.New(apperror.AirPurifierReadFailed, err))
	}

	response := &airpurifierpb.GetCurrentAirPurifierResponse{
		Power:               result.Power,
		IsOn:                result.IsOn,
		Aqi:                 int32(result.AQI),
		AverageAqi:          int32(result.AverageAQI),
		Humidity:            int32(result.Humidity),
		Mode:                string(result.Mode),
		FavoriteLevel:       int32(result.FavoriteLevel),
		FilterLifeRemaining: int32(result.FilterLifeRemaining),
		FilterHoursUsed:     int32(result.FilterHoursUsed),
		MotorSpeed:          int32(result.MotorSpeed),
		PurifyVolume:        int32(result.PurifyVolume),
		Led:                 result.LED,
		ChildLock:           result.ChildLock,
	}

	if result.Temperature != nil {
		response.TemperatureC = wrapperspb.Double(*result.Temperature)
	}

	if result.LEDBrightness != nil {
		response.LedBrightness = wrapperspb.Int32(int32(*result.LEDBrightness))
	}

	if result.Buzzer != nil {
		response.Buzzer = wrapperspb.Bool(*result.Buzzer)
	}

	return response, nil
}
