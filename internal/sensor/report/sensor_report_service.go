package report

import (
	"context"
	"eco-knock-be-embedded/internal/sensor/client"
	"eco-knock-be-embedded/internal/sensor/dto"
	"eco-knock-be-embedded/internal/sensor/streaming"
	"encoding/json"

	"github.com/bytedance/gopkg/util/logger"
)

const DeviceId = 1

type SensorReportService struct {
	client           *client.SensorGRPCClient
	streamingService *streaming.SensorStreamingService
}

func New(
	client *client.SensorGRPCClient,
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
		}

		clientDTO := dto.SensorClientDTO{
			DeviceId: DeviceId,
			Sample:   result.Sample,
		}

		payload, _ := json.Marshal(clientDTO)
		logger.Infof("센서 : %s", payload)
	}
	return nil
}
