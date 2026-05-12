package main

import (
	airservice "eco-knock-be-embedded/internal/airpurifier/xiaomi/service"
	commonconfig "eco-knock-be-embedded/internal/common/config"
	airpurifierpb "eco-knock-be-embedded/internal/grpc/pb/airpurifier/v1"
	lightsensorpb "eco-knock-be-embedded/internal/grpc/pb/lightsensor"
	sensorv1pb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
	sensorv2pb "eco-knock-be-embedded/internal/grpc/pb/sensor/v2"
	airpurifiergrpc "eco-knock-be-embedded/internal/grpc/server/airpurifier"
	lightsensorgrpc "eco-knock-be-embedded/internal/grpc/server/lightsensor"
	sensorgrpc "eco-knock-be-embedded/internal/grpc/server/sensor"
	lightsensorservice "eco-knock-be-embedded/internal/lightsensor/service"
	veml7700reader "eco-knock-be-embedded/internal/lightsensor/veml7700/reader"
	airqualityservice "eco-knock-be-embedded/internal/sensor/airquality/service"
	airqualitystore "eco-knock-be-embedded/internal/sensor/airquality/store"
	bme680reader "eco-knock-be-embedded/internal/sensor/bme680/reader"
	sensorservice "eco-knock-be-embedded/internal/sensor/service"
	"log"
	"net"
	"strconv"

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

	stopLightSensorGRPCServer, err := startLightSensorGRPCServer(grpcServer, commonConfig)
	if err != nil {
		stopSensorGRPCServer()
		_ = listener.Close()
		return nil, err
	}

	stopAirPurifierGRPCServer, err := startAirPurifierGRPCServer(grpcServer, commonConfig)
	if err != nil {
		stopLightSensorGRPCServer()
		stopSensorGRPCServer()
		_ = listener.Close()
		return nil, err
	}

	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Printf("gRPC 서버가 종료되었습니다: %v", err)
		}
	}()

	return func() {
		grpcServer.GracefulStop()
		stopAirPurifierGRPCServer()
		stopLightSensorGRPCServer()
		stopSensorGRPCServer()
		_ = listener.Close()
	}, nil
}

func startSensorGRPCServer(grpcServer *grpc.Server, commonConfig commonconfig.CommonConfig) (func(), error) {
	runtimeConfig, err := resolveSensorRuntimeConfig(commonConfig)
	if err != nil {
		return nil, err
	}

	sensorReader, err := bme680reader.Open(runtimeConfig.readerConfig)
	if err != nil {
		return nil, err
	}

	stateStore, err := airqualitystore.OpenSQLite(runtimeConfig.stateDBPath)
	if err != nil {
		_ = sensorReader.Close()
		return nil, err
	}

	airQualityService, err := airqualityservice.New(runtimeConfig.airQualityConfig)
	if err != nil {
		_ = stateStore.Close()
		_ = sensorReader.Close()
		return nil, err
	}

	sensorQueryService, err := sensorservice.New(
		sensorReader,
		airQualityService,
		stateStore,
		runtimeConfig.serviceConfig,
	)
	if err != nil {
		_ = stateStore.Close()
		_ = sensorReader.Close()
		return nil, err
	}
	if err := sensorQueryService.Start(); err != nil {
		_ = sensorQueryService.Close()
		return nil, err
	}

	sensorV1GRPCServer, err := sensorgrpc.NewV1GRPCServer(sensorQueryService)
	if err != nil {
		_ = sensorQueryService.Close()
		return nil, err
	}

	sensorV2GRPCServer, err := sensorgrpc.NewV2GRPCServer(sensorQueryService)
	if err != nil {
		_ = sensorQueryService.Close()
		return nil, err
	}

	sensorv1pb.RegisterSensorServiceServer(grpcServer, sensorV1GRPCServer)
	sensorv2pb.RegisterSensorServiceServer(grpcServer, sensorV2GRPCServer)

	return func() {
		_ = sensorQueryService.Close()
	}, nil
}

func startLightSensorGRPCServer(grpcServer *grpc.Server, commonConfig commonconfig.CommonConfig) (func(), error) {
	runtimeConfig, err := resolveLightSensorRuntimeConfig(commonConfig)
	if err != nil {
		return nil, err
	}

	lightSensorReader, err := veml7700reader.Open(runtimeConfig.readerConfig)
	if err != nil {
		return nil, err
	}

	lightSensorService, err := lightsensorservice.New(
		lightSensorReader,
		runtimeConfig.readerConfig.PollInterval,
	)
	if err != nil {
		_ = lightSensorReader.Close()
		return nil, err
	}
	if err := lightSensorService.Start(); err != nil {
		_ = lightSensorService.Close()
		return nil, err
	}

	lightSensorGRPCServer, err := lightsensorgrpc.NewGRPCServer(lightSensorService)
	if err != nil {
		_ = lightSensorService.Close()
		return nil, err
	}

	lightsensorpb.RegisterLightSensorServiceServer(grpcServer, lightSensorGRPCServer)

	return func() {
		_ = lightSensorService.Close()
	}, nil
}

func startAirPurifierGRPCServer(grpcServer *grpc.Server, commonConfig commonconfig.CommonConfig) (func(), error) {
	conf, ok, err := resolveAirPurifierConfig(commonConfig)
	if err != nil {
		return nil, err
	}
	if !ok {
		return func() {}, nil
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
