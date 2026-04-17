package service

import (
	airqualityservice "eco-knock-be-embedded/internal/sensor/airquality/service"
	airqualitystore "eco-knock-be-embedded/internal/sensor/airquality/store"
	"eco-knock-be-embedded/internal/sensor/bme680/reader/interfaces"
	sensorconfig "eco-knock-be-embedded/internal/sensor/config"
	"eco-knock-be-embedded/internal/sensor/dto"
	"errors"
	"log"
	"sync"
	"time"
)

const (
	sensorStartMaxAttempts = 3
	sensorStartDelay       = 3 * time.Second
)

var ErrNoSensorSnapshot = errors.New("사용 가능한 센서 스냅샷이 없습니다")

type SensorService struct {
	reader            interfaces.Reader
	airQualityService *airqualityservice.AirQualityService
	stateStore        *airqualitystore.SQLiteStore
	config            sensorconfig.SensorServiceConfig

	mu                    sync.RWMutex
	latestSnapshot        dto.SensorSnapshotDTO
	hasLatestSnapshot     bool
	lastErr               error
	validSamplesSinceSave int

	started  bool
	stopOnce sync.Once
	stopCh   chan struct{}
	doneCh   chan struct{}
}

func New(
	reader interfaces.Reader,
	airQualityService *airqualityservice.AirQualityService,
	stateStore *airqualitystore.SQLiteStore,
	config sensorconfig.SensorServiceConfig,
) (*SensorService, error) {
	if reader == nil {
		return nil, errors.New("센서 리더가 필요합니다")
	}
	if airQualityService == nil {
		return nil, errors.New("공기질 서비스가 필요합니다")
	}
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &SensorService{
		reader:            reader,
		airQualityService: airQualityService,
		stateStore:        stateStore,
		config:            config,
		stopCh:            make(chan struct{}),
		doneCh:            make(chan struct{}),
	}, nil
}

func (service *SensorService) Start() error {
	if service.stateStore != nil {
		state, err := service.stateStore.Load()
		if err != nil {
			return err
		}
		if err := service.airQualityService.Restore(state); err != nil {
			return err
		}
	}

	for attempts := 1; attempts <= sensorStartMaxAttempts; attempts++ {
		if err := service.pollOnce(); err == nil {
			break
		} else {
			log.Printf("센서 시작 실패 -> ( %d / %d ): %v", attempts, sensorStartMaxAttempts, err)
		}

		if attempts < sensorStartMaxAttempts {
			time.Sleep(sensorStartDelay)
		}

		if attempts == sensorStartMaxAttempts {
			log.Printf("초기 센서 스냅샷 없이 서비스를 시작합니다. 다음 폴링 성공을 기다립니다.")
		}
	}

	service.started = true
	go service.pollLoop()
	return nil
}

func (service *SensorService) Read() dto.SensorInternalDTO {
	service.mu.RLock()
	defer service.mu.RUnlock()

	if service.hasLatestSnapshot {
		return dto.SensorInternalDTO{
			Snapshot: service.latestSnapshot,
		}
	}

	if service.lastErr != nil {
		return dto.SensorInternalDTO{Err: service.lastErr}
	}

	return dto.SensorInternalDTO{Err: ErrNoSensorSnapshot}
}

func (service *SensorService) Close() error {
	service.stopOnce.Do(func() {
		close(service.stopCh)
	})
	if service.started {
		<-service.doneCh
	}

	var closeErr error
	if err := service.saveState(); err != nil {
		closeErr = err
	}
	if service.stateStore != nil {
		if err := service.stateStore.Close(); err != nil && closeErr == nil {
			closeErr = err
		}
	}
	if err := service.reader.Close(); err != nil && closeErr == nil {
		closeErr = err
	}

	return closeErr
}

func (service *SensorService) pollLoop() {
	defer close(service.doneCh)

	ticker := time.NewTicker(service.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-service.stopCh:
			return
		case <-ticker.C:
			if err := service.pollOnce(); err != nil {
				log.Printf("센서 폴링 실패: %v", err)
			}
		}
	}
}

func (service *SensorService) pollOnce() error {
	sample, err := service.reader.Read()
	if err != nil {
		service.mu.Lock()
		service.lastErr = err
		service.mu.Unlock()
		return err
	}

	snapshot := dto.SensorSnapshotDTO{
		Sample:     sample,
		AirQuality: service.airQualityService.Estimate(sample),
		MeasuredAt: sample.MeasuredAt,
	}

	service.mu.Lock()
	service.latestSnapshot = snapshot
	service.hasLatestSnapshot = true
	service.lastErr = nil
	service.mu.Unlock()

	if sample.GasValid && sample.HeatStable {
		service.validSamplesSinceSave++
		if service.stateStore != nil &&
			service.validSamplesSinceSave >= service.config.StateCheckpointValidSamples {
			if err := service.saveState(); err != nil {
				return err
			}
			service.validSamplesSinceSave = 0
		}
	}

	return nil
}

func (service *SensorService) saveState() error {
	if service.stateStore == nil {
		return nil
	}

	state := service.airQualityService.State()

	return service.stateStore.Save(state)
}
