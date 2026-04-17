package dto

import (
	"eco-knock-be-embedded/internal/sensor/airquality/dto"
	"time"
)

// SampleDTO Sample은 BME680 강제 측정 한 번의 결과다.
type SampleDTO struct {
	TemperatureC     float64
	HumidityRH       float64
	GasResistanceOhm float64
	Status           uint8
	GasValid         bool
	HeatStable       bool
	MeasuredAt       time.Time
}

// SensorSnapshotDTO SensorSnapshotDTO는 센서 서비스가 발행하는 최신 raw 샘플과 파생값이다.
type SensorSnapshotDTO struct {
	Sample     SampleDTO
	AirQuality dto.AirQualityDTO
	MeasuredAt time.Time
}

// SensorInternalDTO SensorInternalDTO는 센서 서비스가 발행하는 폴링 결과다.
type SensorInternalDTO struct {
	Snapshot SensorSnapshotDTO
	Err      error
}
