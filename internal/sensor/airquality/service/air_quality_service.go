package service

import (
	airqualityconfig "eco-knock-be-embedded/internal/sensor/airquality/config"
	airqualitydto "eco-knock-be-embedded/internal/sensor/airquality/dto"
	airqualitymodel "eco-knock-be-embedded/internal/sensor/airquality/model"
	"eco-knock-be-embedded/internal/sensor/dto"
	"fmt"
	"math"
	"slices"
	"time"
)

const airQualityStateVersion = 1

type AirQualityService struct {
	config                airqualityconfig.AirQualityConfig
	runInStartedAt        time.Time
	learningCompletedAt   time.Time
	totalSampleCount      int
	validSampleCount      int
	lastValidSampleAt     time.Time
	gasBaselineOhm        float64
	humidityReferenceRH   float64
	compensatedGasHistory []float64
	lastOutput            airqualitydto.AirQualityDTO
	hasLastOutput         bool
}

func New(config airqualityconfig.AirQualityConfig) (*AirQualityService, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &AirQualityService{
		config:                config,
		compensatedGasHistory: make([]float64, 0, config.HistoryLimit),
	}, nil
}

func (service *AirQualityService) Restore(state airqualitymodel.AirQualityState) error {
	if state.Version != 0 && state.Version != airQualityStateVersion {
		return fmt.Errorf("지원하지 않는 공기질 상태 버전입니다: %d", state.Version)
	}

	service.runInStartedAt = state.RunInStartedAt
	if state.ValidSampleCount >= service.config.LearningValidSampleGoal && !state.LastOutput.LearningCompleteAt.IsZero() {
		elapsedTarget := state.RunInStartedAt.Add(service.config.LearningDuration)
		service.learningCompletedAt = maxTime(state.LastOutput.LearningCompleteAt, elapsedTarget)
	}
	service.totalSampleCount = max(state.TotalSampleCount, 0)
	service.validSampleCount = max(state.ValidSampleCount, 0)
	service.lastValidSampleAt = state.LastValidSampleAt
	service.gasBaselineOhm = maxFloat(state.GasBaselineOhm, 0)
	service.humidityReferenceRH = clamp(state.HumidityReferenceRH, 0, 100)
	service.compensatedGasHistory = append(
		make([]float64, 0, service.config.HistoryLimit),
		trimHistory(state.CompensatedGasHistory, service.config.HistoryLimit)...,
	)

	if !state.LastOutput.LearningCompleteAt.IsZero() || state.LastOutput.StaticIAQ > 0 {
		service.lastOutput = state.LastOutput
		service.hasLastOutput = true
	}

	if service.gasBaselineOhm == 0 && len(service.compensatedGasHistory) > 0 {
		service.gasBaselineOhm = percentile(service.compensatedGasHistory, 90)
	}

	return nil
}

func (service *AirQualityService) Estimate(sample dto.SampleDTO) airqualitydto.AirQualityDTO {
	if service.runInStartedAt.IsZero() {
		service.runInStartedAt = sample.MeasuredAt
	}

	service.totalSampleCount++
	validSample := sample.GasValid && sample.HeatStable && sample.GasResistanceOhm > 0

	if validSample {
		compensatedGas := compensateGasResistance(sample.GasResistanceOhm, sample.HumidityRH)
		service.compensatedGasHistory = appendTrimmed(
			service.compensatedGasHistory,
			compensatedGas,
			service.config.HistoryLimit,
		)
		service.validSampleCount++
		service.lastValidSampleAt = sample.MeasuredAt
		service.humidityReferenceRH = updateHumidityReference(
			service.humidityReferenceRH,
			sample.HumidityRH,
			service.validSampleCount,
		)
		service.gasBaselineOhm = percentile(service.compensatedGasHistory, 90)
	}

	output := service.deriveOutput(sample, validSample)
	output.Accuracy = service.accuracy(sample.MeasuredAt, validSample)
	output.StabilizationProgressPct = service.stabilizationProgress(sample.MeasuredAt)
	output.LearningCompleteAt = service.learningCompleteAt(sample.MeasuredAt)

	service.lastOutput = output
	service.hasLastOutput = true

	return output
}

func (service *AirQualityService) State() airqualitymodel.AirQualityState {
	history := append([]float64(nil), service.compensatedGasHistory...)

	return airqualitymodel.AirQualityState{
		Version:               airQualityStateVersion,
		RunInStartedAt:        service.runInStartedAt,
		UpdatedAt:             time.Now(),
		TotalSampleCount:      service.totalSampleCount,
		ValidSampleCount:      service.validSampleCount,
		LastValidSampleAt:     service.lastValidSampleAt,
		GasBaselineOhm:        service.gasBaselineOhm,
		HumidityReferenceRH:   service.humidityReferenceRH,
		CompensatedGasHistory: history,
		LastOutput:            service.lastOutput,
	}
}

func (service *AirQualityService) deriveOutput(sample dto.SampleDTO, validSample bool) airqualitydto.AirQualityDTO {
	if !validSample {
		if service.hasLastOutput {
			output := service.lastOutput
			output.Accuracy = minUint32(output.Accuracy, 1)
			return output
		}

		return airqualitydto.AirQualityDTO{
			StaticIAQ:        500,
			EstimatedECO2PPM: 5000,
			EstimatedBVOCPPM: 5,
			GasPercentage:    0,
		}
	}

	compensatedGas := compensateGasResistance(sample.GasResistanceOhm, sample.HumidityRH)
	gasBaseline := service.gasBaselineOhm
	if gasBaseline <= 0 {
		gasBaseline = compensatedGas
	}

	gasRatio := 1.0
	if gasBaseline > 0 {
		gasRatio = clamp(compensatedGas/gasBaseline, 0, 1)
	}

	gasQualityScore := math.Sqrt(gasRatio)
	gasPercentage := gasQualityScore * 100
	humidityScore := humidityComfortScore(sample.HumidityRH)
	airQualityScore := humidityScore + gasQualityScore*75
	staticIAQ := clamp((100-airQualityScore)*5, 0, 500)
	pollutionRatio := staticIAQ / 500

	return airqualitydto.AirQualityDTO{
		StaticIAQ:        staticIAQ,
		EstimatedECO2PPM: clamp(400+4600*math.Pow(pollutionRatio, 2), 400, 5000),
		EstimatedBVOCPPM: clamp(0.05+4.95*math.Pow(pollutionRatio, 1.8), 0.05, 5),
		GasPercentage:    gasPercentage,
	}
}

func (service *AirQualityService) stabilizationProgress(now time.Time) uint32 {
	if service.runInStartedAt.IsZero() {
		return 0
	}

	elapsedProgress := clamp(
		float64(now.Sub(service.runInStartedAt))/float64(service.config.StabilizationDuration),
		0,
		1,
	)
	sampleProgress := clamp(
		float64(service.validSampleCount)/float64(service.config.StabilizationValidSampleGoal),
		0,
		1,
	)

	return uint32(math.Round(minFloat(elapsedProgress, sampleProgress) * 100))
}

func (service *AirQualityService) learningProgress(now time.Time) float64 {
	if service.runInStartedAt.IsZero() {
		return 0
	}

	elapsedProgress := clamp(
		float64(now.Sub(service.runInStartedAt))/float64(service.config.LearningDuration),
		0,
		1,
	)
	sampleProgress := clamp(
		float64(service.validSampleCount)/float64(service.config.LearningValidSampleGoal),
		0,
		1,
	)

	return minFloat(elapsedProgress, sampleProgress)
}

func (service *AirQualityService) accuracy(now time.Time, validSample bool) uint32 {
	stabilized := service.stabilizationProgress(now)
	learningProgress := service.learningProgress(now)

	switch {
	case stabilized < 100 || service.validSampleCount == 0:
		if validSample && service.validSampleCount >= service.config.StabilizationValidSampleGoal/5 {
			return 1
		}
		return 0
	case learningProgress < 0.33:
		if validSample {
			return 1
		}
		return 0
	case learningProgress < 0.8:
		if validSample {
			return 2
		}
		return 1
	default:
		if validSample {
			return 3
		}
		return 2
	}
}

func (service *AirQualityService) learningCompleteAt(now time.Time) time.Time {
	if !service.learningCompletedAt.IsZero() {
		return service.learningCompletedAt
	}

	if service.runInStartedAt.IsZero() {
		return time.Time{}
	}

	elapsedTarget := service.runInStartedAt.Add(service.config.LearningDuration)
	if service.validSampleCount >= service.config.LearningValidSampleGoal {
		if now.Before(elapsedTarget) {
			return elapsedTarget
		}

		if !service.lastOutput.LearningCompleteAt.IsZero() && !service.lastOutput.LearningCompleteAt.After(now) {
			service.learningCompletedAt = maxTime(service.lastOutput.LearningCompleteAt, elapsedTarget)
			return service.learningCompletedAt
		}

		completionAt := service.lastValidSampleAt
		if completionAt.IsZero() || completionAt.Before(elapsedTarget) {
			completionAt = elapsedTarget
		}

		service.learningCompletedAt = completionAt
		return completionAt
	}

	elapsedSeconds := maxFloat(now.Sub(service.runInStartedAt).Seconds(), 1)
	validRatePerSecond := float64(service.validSampleCount) / elapsedSeconds
	if validRatePerSecond <= 0 {
		return elapsedTarget
	}

	remainingSamples := max(service.config.LearningValidSampleGoal-service.validSampleCount, 0)
	sampleTarget := now.Add(time.Duration(float64(remainingSamples)/validRatePerSecond) * time.Second)
	if sampleTarget.Before(elapsedTarget) {
		return elapsedTarget
	}

	return sampleTarget
}

func compensateGasResistance(gasResistance float64, humidityRH float64) float64 {
	humidityOffset := math.Abs(humidityRH - 45)
	penalty := 1 - clamp(humidityOffset/60, 0, 0.4)
	return maxFloat(gasResistance*penalty, 1)
}

func humidityComfortScore(humidityRH float64) float64 {
	switch {
	case humidityRH >= 40 && humidityRH <= 60:
		return 25
	case humidityRH < 40:
		return clamp(25*(humidityRH/40), 0, 25)
	default:
		return clamp(25*((100-humidityRH)/40), 0, 25)
	}
}

func updateHumidityReference(current float64, next float64, validSampleCount int) float64 {
	if validSampleCount <= 1 || current == 0 {
		return next
	}

	alpha := 0.05
	return current + alpha*(next-current)
}

func appendTrimmed(values []float64, next float64, limit int) []float64 {
	values = append(values, next)
	if len(values) <= limit {
		return values
	}

	return append(values[:0], values[len(values)-limit:]...)
}

func trimHistory(values []float64, limit int) []float64 {
	if len(values) <= limit {
		return values
	}

	return values[len(values)-limit:]
}

func percentile(values []float64, pct float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sorted := append([]float64(nil), values...)
	slices.Sort(sorted)

	index := int(math.Round((pct / 100) * float64(len(sorted)-1)))
	index = max(min(index, len(sorted)-1), 0)
	return sorted[index]
}

func clamp(value float64, minValue float64, maxValue float64) float64 {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func maxFloat(left float64, right float64) float64 {
	if left > right {
		return left
	}
	return right
}

func minFloat(left float64, right float64) float64 {
	if left < right {
		return left
	}
	return right
}

func minUint32(left uint32, right uint32) uint32 {
	if left < right {
		return left
	}
	return right
}

func maxTime(left time.Time, right time.Time) time.Time {
	if left.After(right) {
		return left
	}
	return right
}
