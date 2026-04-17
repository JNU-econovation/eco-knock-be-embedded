package config

import "time"

type SensorServiceConfig struct {
	PollInterval                time.Duration
	StateCheckpointValidSamples int
}

func (config SensorServiceConfig) Validate() error {
	switch {
	case config.PollInterval <= 0:
		return invalidConfigError("poll_interval must be greater than zero")
	case config.StateCheckpointValidSamples <= 0:
		return invalidConfigError("state_checkpoint_valid_samples must be greater than zero")
	}

	return nil
}

type invalidConfigError string

func (err invalidConfigError) Error() string {
	return string(err)
}
