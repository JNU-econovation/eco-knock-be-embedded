package model

import (
	"encoding/json"
	"time"

	airqualitydto "eco-knock-be-embedded/internal/sensor/airquality/dto"
)

// AirQualityState AirQualityState는 공기질 추정기의 학습 상태를 저장/복원하기 위한 스냅샷이다.
type AirQualityState struct {
	Version               int
	RunInStartedAt        time.Time
	UpdatedAt             time.Time
	TotalSampleCount      int
	ValidSampleCount      int
	LastValidSampleAt     time.Time
	GasBaselineOhm        float64
	HumidityReferenceRH   float64
	CompensatedGasHistory []float64
	LastOutput            airqualitydto.AirQualityDTO
}

func (state AirQualityState) HistoryJSON() ([]byte, error) {
	return json.Marshal(state.CompensatedGasHistory)
}

func ParseAirQualityHistoryJSON(raw string) ([]float64, error) {
	if raw == "" {
		return nil, nil
	}

	var history []float64
	if err := json.Unmarshal([]byte(raw), &history); err != nil {
		return nil, err
	}

	return history, nil
}
