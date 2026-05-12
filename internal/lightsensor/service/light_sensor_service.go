package service

import (
	"eco-knock-be-embedded/internal/lightsensor/dto"
	"eco-knock-be-embedded/internal/lightsensor/veml7700/reader/interfaces"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	lightSensorStartMaxAttempts = 3
	lightSensorStartDelay       = 3 * time.Second
)

var ErrNoLightSensorSample = errors.New("사용 가능한 조도 센서 샘플이 없습니다")

type LightSensorService struct {
	reader       interfaces.Reader
	pollInterval time.Duration

	mu              sync.RWMutex
	latestSample    dto.SampleDTO
	hasLatestSample bool
	lastErr         error

	started  bool
	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func New(reader interfaces.Reader, pollInterval time.Duration) (*LightSensorService, error) {
	if reader == nil {
		return nil, errors.New("조도 센서 리더가 필요합니다")
	}
	if pollInterval <= 0 {
		return nil, errors.New("조도 센서 polling interval은 0보다 커야 합니다")
	}

	return &LightSensorService{
		reader:       reader,
		pollInterval: pollInterval,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}, nil
}

func (service *LightSensorService) Start() error {
	for attempts := 1; attempts <= lightSensorStartMaxAttempts; attempts++ {
		if err := service.pollOnce(); err == nil {
			break
		} else {
			log.Printf("조도 센서 시작 실패 -> ( %d / %d ): %v", attempts, lightSensorStartMaxAttempts, err)
		}

		if attempts < lightSensorStartMaxAttempts {
			time.Sleep(lightSensorStartDelay)
		}

		if attempts == lightSensorStartMaxAttempts {
			log.Printf("초기 조도 센서 샘플 없이 서비스를 시작합니다. 다음 폴링 성공을 기다립니다.")
		}
	}

	service.started = true
	go service.pollLoop()
	return nil
}

func (service *LightSensorService) Read() dto.LightSensorInternalDTO {
	service.mu.RLock()
	defer service.mu.RUnlock()

	if service.hasLatestSample {
		return dto.LightSensorInternalDTO{
			Sample: service.latestSample,
		}
	}

	if service.lastErr != nil {
		return dto.LightSensorInternalDTO{Err: service.lastErr}
	}

	return dto.LightSensorInternalDTO{Err: ErrNoLightSensorSample}
}

func (service *LightSensorService) Close() error {
	service.stopOnce.Do(func() {
		close(service.stopCh)
	})
	if service.started {
		<-service.doneCh
	}

	return service.reader.Close()
}

func (service *LightSensorService) pollLoop() {
	defer close(service.doneCh)

	ticker := time.NewTicker(service.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-service.stopCh:
			return
		case <-ticker.C:
			if err := service.pollOnce(); err != nil {
				log.Printf("조도 센서 폴링 실패: %v", err)
			}
		}
	}
}

func (service *LightSensorService) pollOnce() error {
	sample, err := service.reader.Read()
	if err != nil {
		service.mu.Lock()
		service.lastErr = err
		service.mu.Unlock()
		return err
	}

	service.mu.Lock()
	service.latestSample = sample
	service.hasLatestSample = true
	service.lastErr = nil
	service.mu.Unlock()

	return nil
}
