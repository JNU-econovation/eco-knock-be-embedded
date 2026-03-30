package streaming

import (
	"context"
	"eco-knock-be-embedded/internal/sensor/bme680/reader/interfaces"
	"errors"
	"time"

	"eco-knock-be-embedded/internal/sensor/dto"
)

const (
	DefaultPollInterval = 3 * time.Second
	streamBufferSize    = 1
)

var ErrReaderRequired = errors.New("sensor reader is required")

// SensorStreamingService SensorStreamingService 폴링 정책만 담당한다.
// Stream 호출마다 전달받은 context에 묶인 별도 결과 채널과 goroutine을 만든다.
type SensorStreamingService struct {
	reader   interfaces.Reader
	interval time.Duration
}

func New(reader interfaces.Reader, interval time.Duration) *SensorStreamingService {
	if interval <= 0 {
		interval = DefaultPollInterval
	}

	return &SensorStreamingService{
		reader:   reader,
		interval: interval,
	}
}

func (service *SensorStreamingService) Stream(ctx context.Context) (<-chan dto.SensorInternalDTO, error) {
	if service.reader == nil {
		return nil, ErrReaderRequired
	}

	stream := make(chan dto.SensorInternalDTO, streamBufferSize)
	go service.run(ctx, stream)

	return stream, nil
}

func (service *SensorStreamingService) PollOnce() dto.SensorInternalDTO {
	sample, err := service.reader.Read()
	return dto.SensorInternalDTO{
		Sample: sample,
		Err:    err,
	}
}

func (service *SensorStreamingService) run(ctx context.Context, stream chan dto.SensorInternalDTO) {
	defer close(stream)

	if !service.publish(ctx, stream, service.PollOnce()) {
		return
	}

	ticker := time.NewTicker(service.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !service.publish(ctx, stream, service.PollOnce()) {
				return
			}
		}
	}
}

func (service *SensorStreamingService) publish(
	ctx context.Context,
	stream chan<- dto.SensorInternalDTO,
	result dto.SensorInternalDTO,
) bool {
	select {
	case <-ctx.Done():
		return false
	case stream <- result:
		return true
	}
}
