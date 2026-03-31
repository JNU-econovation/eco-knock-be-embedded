package report

import (
	"context"
	"eco-knock-be-embedded/internal/grpc/sensor"
	"eco-knock-be-embedded/internal/sensor/dto"
	"eco-knock-be-embedded/internal/sensor/streaming"
	"encoding/json"

	"github.com/bytedance/gopkg/util/logger"
)

const DeviceId = 1

type SensorReportService struct {
	client           *sensor.Client
	streamingService *streaming.SensorStreamingService
}

func New(
	client *sensor.Client,
	streamingService *streaming.SensorStreamingService,
) *SensorReportService {
	return &SensorReportService{
		client:           client,
		streamingService: streamingService,
	}
}

func (service *SensorReportService) Run(ctx context.Context) error {
	stream, err := service.streamingService.Stream(ctx)
	if err != nil {
		return err
	}

	for result := range stream {
		if result.Err != nil {
			logger.Errorf("센서 스트림 중 에러가 발생했습니다 : %v", result.Err)
			return result.Err
		}

		request := dto.SensorClientRequest{
			DeviceId: DeviceId,
			Sample:   result.Sample,
		}

		_, err := service.client.Report(ctx, request)

		if err != nil {
			logger.Errorf("GRPC 요청 중에 에러가 발생했습니다 : %v", err)
			return err
		}
	}
	return nil
}

func (service *SensorReportService) logRequest(request dto.SensorClientRequest) {
	payload, _ := json.Marshal(request)
	logger.Infof("센서 request 객체 : %s", payload)
}
