package main

import (
	commonconfig "eco-knock-be-embedded/internal/common/config"
	sensorpb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
	sensorgrpc "eco-knock-be-embedded/internal/grpc/server/sensor"
	bme680config "eco-knock-be-embedded/internal/sensor/bme680/config"
	bme680reader "eco-knock-be-embedded/internal/sensor/bme680/reader"
	sensorservice "eco-knock-be-embedded/internal/sensor/service"
	"log"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

func startGRPCServer(commonConfig commonconfig.CommonConfig) (func(), error) {
	listener, err := net.Listen("tcp", net.JoinHostPort("", formatPort(commonConfig.ServerGRPCPort)))
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()

	stopSensorGRPCServer, err := startSensorGRPCServer(grpcServer, commonConfig)
	if err != nil {
		_ = listener.Close()
		return nil, err
	}

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("grpc server stopped: %v", err)
		}
	}()

	return func() {
		grpcServer.GracefulStop()
		stopSensorGRPCServer()
		_ = listener.Close()
	}, nil
}

func startSensorGRPCServer(grpcServer *grpc.Server, commonConfig commonconfig.CommonConfig) (func(), error) {
	bme680Conf := bme680config.Config{
		I2CDevice:      commonConfig.SensorI2CDevice,
		I2CAddress:     commonConfig.SensorI2CAddress,
		HeaterTempC:    300,
		HeaterDuration: 100 * time.Millisecond,
		AmbientTempC:   25,
	}
	sensorReader, err := bme680reader.Open(bme680Conf)
	if err != nil {
		return nil, err
	}

	sensorQueryService := sensorservice.New(sensorReader)
	sensorGRPCServer, err := sensorgrpc.NewGRPCServer(sensorQueryService)
	if err != nil {
		_ = sensorReader.Close()
		return nil, err
	}

	sensorpb.RegisterSensorServiceServer(grpcServer, sensorGRPCServer)

	return func() {
		_ = sensorReader.Close()
	}, nil
}

func formatPort(port int) string {
	return strconv.Itoa(port)
}
