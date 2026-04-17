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

var ErrNoSensorSnapshot = errors.New("no sensor snapshot available")

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
		return nil, errors.New("sensor reader is required")
	}
	if airQualityService == nil {
		return nil, errors.New("air quality service is required")
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

	if err := service.pollOnce(); err != nil {
		return err
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
				log.Printf("sensor poll failed: %v", err)
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
