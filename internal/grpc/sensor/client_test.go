package sensor

import (
	"context"
	"errors"
	"testing"
	"time"

	sensorpb "eco-knock-be-embedded/internal/grpc/pb/sensor/v1"
	"eco-knock-be-embedded/internal/sensor/dto"

	"google.golang.org/grpc"
)

func TestNewClientRejectsNilClient(t *testing.T) {
	t.Parallel()

	client, err := NewGRPCClient(nil)
	if !errors.Is(err, ErrClientRequired) {
		t.Fatalf("NewClient() error = %v, want %v", err, ErrClientRequired)
	}
	if client != nil {
		t.Fatal("client != nil, want nil")
	}
}

func TestClientReportMapsSensorDTO(t *testing.T) {
	t.Parallel()

	grpcClient := &fakeSensorServiceClient{}
	client, err := NewGRPCClient(grpcClient)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	measuredAt := time.UnixMilli(1743410400000)
	response, err := client.Report(context.Background(), dto.SensorClientRequest{
		DeviceId: 7,
		Sample: dto.SampleDTO{
			TemperatureC:     24.5,
			HumidityRH:       47.2,
			GasResistanceOhm: 12345.6,
			MeasuredAt:       measuredAt,
		},
	})
	if err != nil {
		t.Fatalf("Report() error = %v", err)
	}

	if response == nil || !response.Accepted {
		t.Fatalf("response = %#v, want accepted response", response)
	}

	if grpcClient.request == nil {
		t.Fatal("request = nil, want request")
	}
	if grpcClient.request.DeviceId != 7 {
		t.Fatalf("DeviceId = %v, want 7", grpcClient.request.DeviceId)
	}
	if grpcClient.request.TemperatureC != 24.5 {
		t.Fatalf("TemperatureC = %v, want 24.5", grpcClient.request.TemperatureC)
	}
	if grpcClient.request.HumidityRh != 47.2 {
		t.Fatalf("HumidityRh = %v, want 47.2", grpcClient.request.HumidityRh)
	}
	if grpcClient.request.GasResistanceOhm != 12345.6 {
		t.Fatalf("GasResistanceOhm = %v, want 12345.6", grpcClient.request.GasResistanceOhm)
	}
	if grpcClient.request.MeasuredAtUnixMs != measuredAt.UnixMilli() {
		t.Fatalf("MeasuredAtUnixMs = %v, want %v", grpcClient.request.MeasuredAtUnixMs, measuredAt.UnixMilli())
	}
}

type fakeSensorServiceClient struct {
	request *sensorpb.ReportSensorRequest
}

func (client *fakeSensorServiceClient) ReportSensor(
	_ context.Context,
	request *sensorpb.ReportSensorRequest,
	_ ...grpc.CallOption,
) (*sensorpb.ReportSensorResponse, error) {
	client.request = request
	return &sensorpb.ReportSensorResponse{
		Accepted: true,
	}, nil
}
