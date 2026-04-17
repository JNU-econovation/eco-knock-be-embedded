package main

import (
	commonconfig "eco-knock-be-embedded/internal/common/config"
	airqualityconfig "eco-knock-be-embedded/internal/sensor/airquality/config"
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	sensorconfig "eco-knock-be-embedded/internal/sensor/config"
)

type sensorRuntimeConfig struct {
	readerConfig     bme680config.Config
	airQualityConfig airqualityconfig.AirQualityConfig
	serviceConfig    sensorconfig.SensorServiceConfig
	stateDBPath      string
}

func resolveSensorRuntimeConfig(commonConfig commonconfig.CommonConfig) (sensorRuntimeConfig, error) {
	readerConfig := bme680config.Config{
		I2CDevice:      commonConfig.SensorI2CDevice,
		I2CAddress:     commonConfig.SensorI2CAddress,
		HeaterTempC:    commonConfig.SensorHeaterTempC,
		HeaterDuration: commonConfig.SensorHeaterDuration,
		AmbientTempC:   commonConfig.SensorAmbientTempC,
	}
	if err := readerConfig.Validate(); err != nil {
		return sensorRuntimeConfig{}, err
	}

	airQualityConfig := airqualityconfig.AirQualityConfig{
		HistoryLimit:                 commonConfig.SensorAirQualityHistoryLimit,
		StabilizationDuration:        commonConfig.SensorAirQualityStabilizationDuration,
		LearningDuration:             commonConfig.SensorAirQualityLearningDuration,
		StabilizationValidSampleGoal: commonConfig.SensorAirQualityStabilizationValidSampleGoal,
		LearningValidSampleGoal:      commonConfig.SensorAirQualityLearningValidSampleGoal,
	}
	if err := airQualityConfig.Validate(); err != nil {
		return sensorRuntimeConfig{}, err
	}

	serviceConfig := sensorconfig.SensorServiceConfig{
		PollInterval:                commonConfig.SensorPollInterval,
		StateCheckpointValidSamples: commonConfig.SensorStateCheckpointValidSamples,
	}
	if err := serviceConfig.Validate(); err != nil {
		return sensorRuntimeConfig{}, err
	}

	return sensorRuntimeConfig{
		readerConfig:     readerConfig,
		airQualityConfig: airQualityConfig,
		serviceConfig:    serviceConfig,
		stateDBPath:      commonConfig.SensorStateDBPath,
	}, nil
}
