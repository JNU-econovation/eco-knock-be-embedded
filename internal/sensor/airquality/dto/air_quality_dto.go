package dto

import "time"

// AirQualityDTO AirQualityDTO는 raw 샘플에서 추정한 실내 공기질 파생값이다.
type AirQualityDTO struct {
	StaticIAQ                float64
	EstimatedECO2PPM         float64
	EstimatedBVOCPPM         float64
	Accuracy                 uint32
	StabilizationProgressPct uint32
	GasPercentage            float64
	LearningCompleteAt       time.Time
}
