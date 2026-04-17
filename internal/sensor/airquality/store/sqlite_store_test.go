package store

import (
	airqualitydto "eco-knock-be-embedded/internal/sensor/airquality/dto"
	airqualitymodel "eco-knock-be-embedded/internal/sensor/airquality/model"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

func TestSQLiteStoreSaveAndLoad(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "sensor_state.db")
	store, err := OpenSQLite(dbPath)
	if err != nil {
		t.Fatalf("open sqlite store: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	expected := airqualitymodel.AirQualityState{
		Version:             1,
		RunInStartedAt:      time.Unix(1_700_000_000, 0),
		UpdatedAt:           time.Unix(1_700_000_600, 0),
		TotalSampleCount:    4200,
		ValidSampleCount:    3600,
		LastValidSampleAt:   time.Unix(1_700_000_599, 0),
		GasBaselineOhm:      124_532.42,
		HumidityReferenceRH: 46.3,
		CompensatedGasHistory: []float64{
			118_000,
			120_500,
			122_100,
		},
		LastOutput: airqualitydto.AirQualityDTO{
			StaticIAQ:                42.5,
			EstimatedECO2PPM:         812,
			EstimatedBVOCPPM:         0.74,
			Accuracy:                 3,
			StabilizationProgressPct: 100,
			GasPercentage:            82.1,
			LearningCompleteAt:       time.Unix(1_700_003_600, 0),
		},
	}

	if err := store.Save(expected); err != nil {
		t.Fatalf("save sqlite state: %v", err)
	}

	loaded, err := store.Load()
	if err != nil {
		t.Fatalf("load sqlite state: %v", err)
	}

	if loaded.Version != expected.Version {
		t.Fatalf("expected version %d, got %d", expected.Version, loaded.Version)
	}
	if !loaded.RunInStartedAt.Equal(expected.RunInStartedAt) {
		t.Fatalf("expected run in start %v, got %v", expected.RunInStartedAt, loaded.RunInStartedAt)
	}
	if loaded.TotalSampleCount != expected.TotalSampleCount {
		t.Fatalf("expected total samples %d, got %d", expected.TotalSampleCount, loaded.TotalSampleCount)
	}
	if loaded.ValidSampleCount != expected.ValidSampleCount {
		t.Fatalf("expected valid samples %d, got %d", expected.ValidSampleCount, loaded.ValidSampleCount)
	}
	if !loaded.LastValidSampleAt.Equal(expected.LastValidSampleAt) {
		t.Fatalf("expected last valid sample %v, got %v", expected.LastValidSampleAt, loaded.LastValidSampleAt)
	}
	if loaded.GasBaselineOhm != expected.GasBaselineOhm {
		t.Fatalf("expected baseline %.2f, got %.2f", expected.GasBaselineOhm, loaded.GasBaselineOhm)
	}
	if loaded.HumidityReferenceRH != expected.HumidityReferenceRH {
		t.Fatalf("expected humidity reference %.2f, got %.2f", expected.HumidityReferenceRH, loaded.HumidityReferenceRH)
	}
	if !slices.Equal(loaded.CompensatedGasHistory, expected.CompensatedGasHistory) {
		t.Fatalf("expected compensated history %v, got %v", expected.CompensatedGasHistory, loaded.CompensatedGasHistory)
	}
	if loaded.LastOutput.StaticIAQ != expected.LastOutput.StaticIAQ {
		t.Fatalf("expected static IAQ %.2f, got %.2f", expected.LastOutput.StaticIAQ, loaded.LastOutput.StaticIAQ)
	}
	if loaded.LastOutput.EstimatedECO2PPM != expected.LastOutput.EstimatedECO2PPM {
		t.Fatalf("expected eCO2 %.2f, got %.2f", expected.LastOutput.EstimatedECO2PPM, loaded.LastOutput.EstimatedECO2PPM)
	}
	if loaded.LastOutput.EstimatedBVOCPPM != expected.LastOutput.EstimatedBVOCPPM {
		t.Fatalf("expected bVOC %.2f, got %.2f", expected.LastOutput.EstimatedBVOCPPM, loaded.LastOutput.EstimatedBVOCPPM)
	}
	if loaded.LastOutput.Accuracy != expected.LastOutput.Accuracy {
		t.Fatalf("expected accuracy %d, got %d", expected.LastOutput.Accuracy, loaded.LastOutput.Accuracy)
	}
	if loaded.LastOutput.StabilizationProgressPct != expected.LastOutput.StabilizationProgressPct {
		t.Fatalf(
			"expected stabilization progress %d, got %d",
			expected.LastOutput.StabilizationProgressPct,
			loaded.LastOutput.StabilizationProgressPct,
		)
	}
	if loaded.LastOutput.GasPercentage != expected.LastOutput.GasPercentage {
		t.Fatalf("expected gas percentage %.2f, got %.2f", expected.LastOutput.GasPercentage, loaded.LastOutput.GasPercentage)
	}
	if !loaded.LastOutput.LearningCompleteAt.Equal(expected.LastOutput.LearningCompleteAt) {
		t.Fatalf(
			"expected learning complete %v, got %v",
			expected.LastOutput.LearningCompleteAt,
			loaded.LastOutput.LearningCompleteAt,
		)
	}
}
