package dto

import (
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

// SensorInternalDTO SensorDTO는 센서 서비스가 발행하는 폴링 결과다.
type SensorInternalDTO struct {
	Sample SampleDTO
	Err    error
}

type SensorClientDTO struct {
	DeviceId int
	Sample   SampleDTO
}
