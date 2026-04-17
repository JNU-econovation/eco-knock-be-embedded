package store

import (
	"database/sql"
	airqualitymodel "eco-knock-be-embedded/internal/sensor/airquality/model"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

const defaultSensorID = "bme680-default"

type SQLiteStore struct {
	db       *sql.DB
	sensorID string
}

func OpenSQLite(path string) (*SQLiteStore, error) {
	if path == "" {
		return nil, errors.New("sqlite path is required")
	}

	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{
		db:       db,
		sensorID: defaultSensorID,
	}

	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (store *SQLiteStore) Load() (airqualitymodel.AirQualityState, error) {
	const query = `
SELECT
	version,
	run_in_started_at_unix_ms,
	updated_at_unix_ms,
	total_sample_count,
	valid_sample_count,
	last_valid_sample_at_unix_ms,
	gas_baseline_ohm,
	humidity_reference_rh,
	compensated_gas_history_json,
	last_static_iaq,
	last_estimated_eco2_ppm,
	last_estimated_bvoc_ppm,
	last_accuracy,
	last_stabilization_progress_pct,
	last_gas_percentage,
	learning_complete_at_unix_ms
FROM sensor_air_quality_state
WHERE sensor_id = ?`

	var (
		state                     airqualitymodel.AirQualityState
		runInStartedAtUnixMs      int64
		updatedAtUnixMs           int64
		lastValidSampleAtUnixMs   int64
		historyJSON               string
		learningCompleteAtUnixMs  int64
		lastAccuracy              int64
		lastStabilizationProgress int64
	)

	err := store.db.QueryRow(query, store.sensorID).Scan(
		&state.Version,
		&runInStartedAtUnixMs,
		&updatedAtUnixMs,
		&state.TotalSampleCount,
		&state.ValidSampleCount,
		&lastValidSampleAtUnixMs,
		&state.GasBaselineOhm,
		&state.HumidityReferenceRH,
		&historyJSON,
		&state.LastOutput.StaticIAQ,
		&state.LastOutput.EstimatedECO2PPM,
		&state.LastOutput.EstimatedBVOCPPM,
		&lastAccuracy,
		&lastStabilizationProgress,
		&state.LastOutput.GasPercentage,
		&learningCompleteAtUnixMs,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return airqualitymodel.AirQualityState{}, nil
	}
	if err != nil {
		return airqualitymodel.AirQualityState{}, err
	}

	history, err := airqualitymodel.ParseAirQualityHistoryJSON(historyJSON)
	if err != nil {
		return airqualitymodel.AirQualityState{}, err
	}

	state.RunInStartedAt = unixMilli(runInStartedAtUnixMs)
	state.UpdatedAt = unixMilli(updatedAtUnixMs)
	state.LastValidSampleAt = unixMilli(lastValidSampleAtUnixMs)
	state.LastOutput.Accuracy = uint32(lastAccuracy)
	state.LastOutput.StabilizationProgressPct = uint32(lastStabilizationProgress)
	state.LastOutput.LearningCompleteAt = unixMilli(learningCompleteAtUnixMs)
	state.CompensatedGasHistory = history

	return state, nil
}

func (store *SQLiteStore) Save(state airqualitymodel.AirQualityState) error {
	historyJSON, err := state.HistoryJSON()
	if err != nil {
		return err
	}

	const statement = `
INSERT INTO sensor_air_quality_state (
	sensor_id,
	version,
	run_in_started_at_unix_ms,
	updated_at_unix_ms,
	total_sample_count,
	valid_sample_count,
	last_valid_sample_at_unix_ms,
	gas_baseline_ohm,
	humidity_reference_rh,
	compensated_gas_history_json,
	last_static_iaq,
	last_estimated_eco2_ppm,
	last_estimated_bvoc_ppm,
	last_accuracy,
	last_stabilization_progress_pct,
	last_gas_percentage,
	learning_complete_at_unix_ms
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(sensor_id) DO UPDATE SET
	version = excluded.version,
	run_in_started_at_unix_ms = excluded.run_in_started_at_unix_ms,
	updated_at_unix_ms = excluded.updated_at_unix_ms,
	total_sample_count = excluded.total_sample_count,
	valid_sample_count = excluded.valid_sample_count,
	last_valid_sample_at_unix_ms = excluded.last_valid_sample_at_unix_ms,
	gas_baseline_ohm = excluded.gas_baseline_ohm,
	humidity_reference_rh = excluded.humidity_reference_rh,
	compensated_gas_history_json = excluded.compensated_gas_history_json,
	last_static_iaq = excluded.last_static_iaq,
	last_estimated_eco2_ppm = excluded.last_estimated_eco2_ppm,
	last_estimated_bvoc_ppm = excluded.last_estimated_bvoc_ppm,
	last_accuracy = excluded.last_accuracy,
	last_stabilization_progress_pct = excluded.last_stabilization_progress_pct,
	last_gas_percentage = excluded.last_gas_percentage,
	learning_complete_at_unix_ms = excluded.learning_complete_at_unix_ms`

	_, err = store.db.Exec(
		statement,
		store.sensorID,
		state.Version,
		toUnixMilli(state.RunInStartedAt),
		toUnixMilli(nonZeroTime(state.UpdatedAt)),
		state.TotalSampleCount,
		state.ValidSampleCount,
		toUnixMilli(state.LastValidSampleAt),
		state.GasBaselineOhm,
		state.HumidityReferenceRH,
		string(historyJSON),
		state.LastOutput.StaticIAQ,
		state.LastOutput.EstimatedECO2PPM,
		state.LastOutput.EstimatedBVOCPPM,
		state.LastOutput.Accuracy,
		state.LastOutput.StabilizationProgressPct,
		state.LastOutput.GasPercentage,
		toUnixMilli(state.LastOutput.LearningCompleteAt),
	)

	return err
}

func (store *SQLiteStore) Close() error {
	if store == nil || store.db == nil {
		return nil
	}

	return store.db.Close()
}

func (store *SQLiteStore) init() error {
	pragmas := []string{
		"PRAGMA journal_mode = WAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, statement := range pragmas {
		if _, err := store.db.Exec(statement); err != nil {
			return err
		}
	}

	const schema = `
CREATE TABLE IF NOT EXISTS sensor_air_quality_state (
	sensor_id TEXT PRIMARY KEY,
	version INTEGER NOT NULL,
	run_in_started_at_unix_ms INTEGER NOT NULL,
	updated_at_unix_ms INTEGER NOT NULL,
	total_sample_count INTEGER NOT NULL,
	valid_sample_count INTEGER NOT NULL,
	last_valid_sample_at_unix_ms INTEGER NOT NULL,
	gas_baseline_ohm REAL NOT NULL,
	humidity_reference_rh REAL NOT NULL,
	compensated_gas_history_json TEXT NOT NULL,
	last_static_iaq REAL NOT NULL,
	last_estimated_eco2_ppm REAL NOT NULL,
	last_estimated_bvoc_ppm REAL NOT NULL,
	last_accuracy INTEGER NOT NULL,
	last_stabilization_progress_pct INTEGER NOT NULL,
	last_gas_percentage REAL NOT NULL,
	learning_complete_at_unix_ms INTEGER NOT NULL
)`

	if _, err := store.db.Exec(schema); err != nil {
		return fmt.Errorf("create sensor_air_quality_state: %w", err)
	}

	return nil
}

func nonZeroTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now()
	}
	return value
}

func toUnixMilli(value time.Time) int64 {
	if value.IsZero() {
		return 0
	}
	return value.UnixMilli()
}

func unixMilli(value int64) time.Time {
	if value == 0 {
		return time.Time{}
	}
	return time.UnixMilli(value)
}
