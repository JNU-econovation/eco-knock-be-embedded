package main

import (
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	airservice "eco-knock-be-embedded/internal/airpurifier/xiaomi/service"
	commonconfig "eco-knock-be-embedded/internal/common/config"
	airpurifierpb "eco-knock-be-embedded/internal/grpc/pb/airpurifier/v1"
	sensorpb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
	airpurifiergrpc "eco-knock-be-embedded/internal/grpc/server/airpurifier"
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

	stopAirPurifierGRPCServer, err := startAirPurifierGRPCServer(grpcServer, commonConfig)
	if err != nil {
		stopSensorGRPCServer()
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
		stopAirPurifierGRPCServer()
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

func startAirPurifierGRPCServer(grpcServer *grpc.Server, commonConfig commonconfig.CommonConfig) (func(), error) {
	if commonConfig.AirPurifierAddress == "" || commonConfig.AirPurifierToken == "" || commonConfig.AirPurifierTimeout <= 0 {
		log.Printf("air purifier grpc server skipped: configuration is incomplete")
		return func() {}, nil
	}

	conf, err := airconfig.New(commonConfig.AirPurifierAddress, commonConfig.AirPurifierToken, commonConfig.AirPurifierTimeout)
	if err != nil {
		return nil, err
	}

	airPurifierService, err := airservice.New(conf)
	if err != nil {
		return nil, err
	}

	airPurifierGRPCServer, err := airpurifiergrpc.NewGRPCServer(airPurifierService)
	if err != nil {
		return nil, err
	}

	airpurifierpb.RegisterAirPurifierServiceServer(grpcServer, airPurifierGRPCServer)
	return func() {}, nil
}

func formatPort(port int) string {
	return strconv.Itoa(port)
}
