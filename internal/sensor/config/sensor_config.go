package config

import "time"

type SensorServiceConfig struct {
	PollInterval                time.Duration
	StateCheckpointValidSamples int
}

func (config SensorServiceConfig) Validate() error {
	switch {
	case config.PollInterval <= 0:
		return invalidConfigError("poll_interval은 0보다 커야 합니다")
	case config.StateCheckpointValidSamples <= 0:
		return invalidConfigError("state_checkpoint_valid_samples는 0보다 커야 합니다")
	}

	return nil
}

type invalidConfigError string

func (err invalidConfigError) Error() string {
	return string(err)
}
