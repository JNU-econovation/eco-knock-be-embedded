package main

import (
	"context"
	"eco-knock-be-embedded/internal/common/config"
	"eco-knock-be-embedded/internal/grpc/common"
	sensorpb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
	"eco-knock-be-embedded/internal/grpc/sensor"
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	bme680reader "eco-knock-be-embedded/internal/sensor/bme680/reader"
	"eco-knock-be-embedded/internal/sensor/report"
	"eco-knock-be-embedded/internal/sensor/streaming"
	"log"
	"time"
)

func startSensorReporter(commonConfig config.CommonConfig) (func(), error) {
	if commonConfig.CentralBackendURL == "" && commonConfig.AllowCentralBackendFailure {
		log.Println("central backend address is empty; sensor reporter will be skipped")
		return func() {}, nil
	}

	grpcConf := common.GrpcConfig{
		Address: commonConfig.CentralBackendURL,
		Timeout: 5 * time.Second,
	}

	conn, err := common.NewGRPCConnect(context.Background(), grpcConf)
	if err != nil {
		if commonConfig.AllowCentralBackendFailure {
			log.Printf("central backend connection skipped: %v", err)
			return func() {}, nil
		}

		return nil, err
	}

	rawClient := sensorpb.NewSensorServiceClient(conn)

	sensorGRPCClient, err := sensor.NewGRPCClient(rawClient)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}

	bme680Conf := bme680config.Config{
		I2CDevice:      "/dev/i2c-1",
		I2CAddress:     0x76,
		HeaterTempC:    300,
		HeaterDuration: 100 * time.Millisecond,
		AmbientTempC:   25,
	}
	sensorReader, err := bme680reader.Open(bme680Conf)

	if err != nil {
		_ = conn.Close()
		return nil, err
	}

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
		_ = conn.Close()
	}, nil
}
