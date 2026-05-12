package main

import (
	commonconfig "eco-knock-be-embedded/internal/common/config"
	veml7700config "eco-knock-be-embedded/internal/lightsensor/veml7700/config"
)

type lightSensorRuntimeConfig struct {
	readerConfig veml7700config.Config
}

func resolveLightSensorRuntimeConfig(commonConfig commonconfig.CommonConfig) (lightSensorRuntimeConfig, error) {
	readerConfig := veml7700config.Config{
		I2CDevice:    commonConfig.LightSensorI2CDevice,
		I2CAddress:   commonConfig.LightSensorI2CAddress,
		PollInterval: commonConfig.LightSensorPollInterval,
	}
	if err := readerConfig.Validate(); err != nil {
		return lightSensorRuntimeConfig{}, err
	}

	return lightSensorRuntimeConfig{
		readerConfig: readerConfig,
	}, nil
}
