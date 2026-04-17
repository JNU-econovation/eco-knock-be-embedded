package service

import (
	"eco-knock-be-embedded/internal/sensor/airquality/config"
	airqualitydto "eco-knock-be-embedded/internal/sensor/airquality/dto"
	"eco-knock-be-embedded/internal/sensor/dto"
	"testing"
	"time"
)

func testAirQualityConfig() config.AirQualityConfig {
	return config.AirQualityConfig{
		HistoryLimit:                 3600,
		StabilizationDuration:        5 * time.Minute,
		LearningDuration:             time.Hour,
		StabilizationValidSampleGoal: 300,
		LearningValidSampleGoal:      3600,
	}
}

func TestAirQualityServiceProducesStableOutputsAfterLearning(t *testing.T) {
	service, err := New(testAirQualityConfig())
	if err != nil {
		t.Fatalf("new air quality service: %v", err)
	}
	startedAt := time.Unix(1_700_000_000, 0)

	var output airqualitydto.AirQualityDTO
	for i := 0; i < 3600; i++ {
		output = service.Estimate(dto.SampleDTO{
			TemperatureC:     24.5,
			HumidityRH:       45,
			GasResistanceOhm: 120_000,
			GasValid:         true,
			HeatStable:       true,
			MeasuredAt:       startedAt.Add(time.Duration(i) * time.Second),
		})
	}

	if output.Accuracy != 3 {
		t.Fatalf("expected accuracy 3, got %d", output.Accuracy)
	}
	if output.StabilizationProgressPct != 100 {
		t.Fatalf("expected stabilization 100, got %d", output.StabilizationProgressPct)
	}
	if output.StaticIAQ > 10 {
		t.Fatalf("expected low static IAQ for clean baseline, got %.2f", output.StaticIAQ)
	}
	if output.EstimatedECO2PPM > 450 {
		t.Fatalf("expected low estimated eCO2 for clean baseline, got %.2f", output.EstimatedECO2PPM)
	}
	if output.GasPercentage < 99 {
		t.Fatalf("expected gas percentage near 100, got %.2f", output.GasPercentage)
	}
	if output.LearningCompleteAt.Before(startedAt.Add(time.Hour)) {
		t.Fatalf("expected learning completion after one hour, got %v", output.LearningCompleteAt)
	}
}

func TestAirQualityServicePollutionSpikeWorsensDerivedValues(t *testing.T) {
	service, err := New(testAirQualityConfig())
	if err != nil {
		t.Fatalf("new air quality service: %v", err)
	}
	startedAt := time.Unix(1_700_010_000, 0)

	for i := 0; i < 3600; i++ {
		service.Estimate(dto.SampleDTO{
			TemperatureC:     24.5,
			HumidityRH:       45,
			GasResistanceOhm: 120_000,
			GasValid:         true,
			HeatStable:       true,
			MeasuredAt:       startedAt.Add(time.Duration(i) * time.Second),
		})
	}

	clean := service.Estimate(dto.SampleDTO{
		TemperatureC:     24.5,
		HumidityRH:       45,
		GasResistanceOhm: 120_000,
		GasValid:         true,
		HeatStable:       true,
		MeasuredAt:       startedAt.Add(time.Hour),
	})
	polluted := service.Estimate(dto.SampleDTO{
		TemperatureC:     25.1,
		HumidityRH:       68,
		GasResistanceOhm: 24_000,
		GasValid:         true,
		HeatStable:       true,
		MeasuredAt:       startedAt.Add(time.Hour + time.Second),
	})

	if polluted.StaticIAQ <= clean.StaticIAQ {
		t.Fatalf("expected polluted IAQ %.2f to be worse than clean IAQ %.2f", polluted.StaticIAQ, clean.StaticIAQ)
	}
	if polluted.EstimatedECO2PPM <= clean.EstimatedECO2PPM {
		t.Fatalf(
			"expected polluted eCO2 %.2f to be higher than clean eCO2 %.2f",
			polluted.EstimatedECO2PPM,
			clean.EstimatedECO2PPM,
		)
	}
	if polluted.EstimatedBVOCPPM <= clean.EstimatedBVOCPPM {
		t.Fatalf(
			"expected polluted bVOC %.2f to be higher than clean bVOC %.2f",
			polluted.EstimatedBVOCPPM,
			clean.EstimatedBVOCPPM,
		)
	}
	if polluted.GasPercentage >= clean.GasPercentage {
		t.Fatalf(
			"expected polluted gas percentage %.2f to be lower than clean %.2f",
			polluted.GasPercentage,
			clean.GasPercentage,
		)
	}
}

func TestAirQualityServiceKeepsLastOutputWhenSampleIsInvalid(t *testing.T) {
	service, err := New(testAirQualityConfig())
	if err != nil {
		t.Fatalf("new air quality service: %v", err)
	}
	startedAt := time.Unix(1_700_020_000, 0)

	var learned airqualitydto.AirQualityDTO
	for i := 0; i < 3600; i++ {
		learned = service.Estimate(dto.SampleDTO{
			TemperatureC:     24.5,
			HumidityRH:       45,
			GasResistanceOhm: 120_000,
			GasValid:         true,
			HeatStable:       true,
			MeasuredAt:       startedAt.Add(time.Duration(i) * time.Second),
		})
	}

	invalid := service.Estimate(dto.SampleDTO{
		TemperatureC:     24.8,
		HumidityRH:       44,
		GasResistanceOhm: 0,
		GasValid:         false,
		HeatStable:       false,
		MeasuredAt:       startedAt.Add(time.Hour + time.Second),
	})

	if invalid.StaticIAQ != learned.StaticIAQ {
		t.Fatalf("expected invalid sample to keep last static IAQ %.2f, got %.2f", learned.StaticIAQ, invalid.StaticIAQ)
	}
	if invalid.EstimatedECO2PPM != learned.EstimatedECO2PPM {
		t.Fatalf(
			"expected invalid sample to keep last estimated eCO2 %.2f, got %.2f",
			learned.EstimatedECO2PPM,
			invalid.EstimatedECO2PPM,
		)
	}
	if invalid.Accuracy >= learned.Accuracy {
		t.Fatalf("expected invalid sample accuracy %d to drop below learned accuracy %d", invalid.Accuracy, learned.Accuracy)
	}
}
