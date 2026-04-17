package config

import "time"

type AirQualityConfig struct {
	HistoryLimit                 int
	StabilizationDuration        time.Duration
	LearningDuration             time.Duration
	StabilizationValidSampleGoal int
	LearningValidSampleGoal      int
}

func (config AirQualityConfig) Validate() error {
	switch {
	case config.HistoryLimit <= 0:
		return invalidConfigError("history_limit must be greater than zero")
	case config.StabilizationDuration <= 0:
		return invalidConfigError("stabilization_duration must be greater than zero")
	case config.LearningDuration <= 0:
		return invalidConfigError("learning_duration must be greater than zero")
	case config.StabilizationValidSampleGoal <= 0:
		return invalidConfigError("stabilization_valid_sample_goal must be greater than zero")
	case config.LearningValidSampleGoal <= 0:
		return invalidConfigError("learning_valid_sample_goal must be greater than zero")
	}

	return nil
}

type invalidConfigError string

func (err invalidConfigError) Error() string {
	return string(err)
}
