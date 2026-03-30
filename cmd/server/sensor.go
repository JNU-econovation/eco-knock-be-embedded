package main

import (
	"context"
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	bme680reader "eco-knock-be-embedded/internal/sensor/bme680/reader"
	"eco-knock-be-embedded/internal/sensor/client"
	"eco-knock-be-embedded/internal/sensor/report"
	"eco-knock-be-embedded/internal/sensor/streaming"
	"log"
	"time"
)

func startSensorReporter() (func(), error) {
	conf := bme680config.Config{
		I2CDevice:      "/dev/i2c-1",
		I2CAddress:     0x76,
		HeaterTempC:    300,
		HeaterDuration: 100 * time.Millisecond,
		AmbientTempC:   25,
	}
	sensorReader, err := bme680reader.Open(conf)

	if err != nil {
		return nil, err
	}

	sensorGRPCClient := client.New()
	sensorStreamingService := streaming.New(sensorReader, 0)
	sensorReportService := report.New(sensorGRPCClient, sensorStreamingService)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		if err := sensorReportService.Run(ctx); err != nil {
			log.Printf("sensor report service stopped: %v", err)
			cancel()
		}
	}()

	return func() {
		cancel()
		_ = sensorReader.Close()
	}, nil
}
