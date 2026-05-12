package dto

import "time"

// SampleDTO SampleDTO는 VEML7700 조도 측정 한 번의 결과다.
type SampleDTO struct {
	Lux        float64
	RawALS     uint16
	RawWhite   uint16
	MeasuredAt time.Time
}

// LightSensorInternalDTO LightSensorInternalDTO는 조도 센서 서비스가 발행하는 폴링 결과다.
type LightSensorInternalDTO struct {
	Sample SampleDTO
	Err    error
}
