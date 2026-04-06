package test

import (
	"context"
	"crypto/md5"
	airconfig "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/constant"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/service"
	testclient "eco-knock-be-embedded/internal/airpurifier/xiaomi/test/client"
	"encoding/binary"
	"encoding/json"
	"errors"
	"testing"
	"time"

	miiorequest "eco-knock-be-embedded/internal/airpurifier/xiaomi/dto/miio/request"
	miioresponse "eco-knock-be-embedded/internal/airpurifier/xiaomi/dto/miio/response"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/miio"
)

func TestNewRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	conf, err := airconfig.New("192.168.0.10", "abcd", time.Second)
	if !errors.Is(err, constant.ErrInvalidTokenLength) {
		t.Fatalf("expected ErrInvalidTokenLength, got %v", err)
	}

	if conf.Address != "" || conf.Token != nil || conf.Timeout != 0 {
		t.Fatalf("expected zero config, got %+v", conf)
	}
}

func TestStatusSendsHelloThenGetProperties(t *testing.T) {
	t.Parallel()

	token := "00112233445566778899aabbccddeeff"
	conf, err := airconfig.New("192.168.0.10", token, time.Second)
	if err != nil {
		t.Fatalf("unexpected new config error: %v", err)
	}

	deviceID := [4]byte{0x11, 0x22, 0x33, 0x44}
	client := testclient.NewFakeClient(
		buildHelloResponse(deviceID, 100),
		buildRPCResponse(t, token, deviceID, 101, miioresponse.MIIOResponse{
			ID: 1,
			Result: mustMarshalJSON(t, []any{
				"on",
				12,
				10,
				48,
				231,
				"favorite",
				7,
				82,
				400,
				755,
				12345,
				"on",
				1,
				"off",
				"on",
			}),
		}),
	)
	airPurifierService, err := service.NewWithClient(service.Config{
		Address: conf.Address,
		Token:   conf.Token,
		Timeout: conf.Timeout,
	}, client)
	if err != nil {
		t.Fatalf("unexpected new service error: %v", err)
	}

	status, err := airPurifierService.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected status error: %v", err)
	}

	if !status.IsOn {
		t.Fatal("expected purifier to be on")
	}

	if status.Mode != constant.OperationModeFavorite {
		t.Fatalf("expected mode favorite, got %s", status.Mode)
	}

	if status.FavoriteLevel != 7 {
		t.Fatalf("expected favorite level 7, got %d", status.FavoriteLevel)
	}

	if status.Temperature == nil || *status.Temperature != 23.1 {
		t.Fatalf("expected temperature 23.1, got %#v", status.Temperature)
	}

	requests := client.Requests()
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}

	request := decodeRequestPayload(t, requests[1], token)
	if request.Method != string(constant.MIIOMethodGetProp) {
		t.Fatalf("expected %s method, got %s", constant.MIIOMethodGetProp, request.Method)
	}

	properties, ok := request.Params.([]any)
	if !ok {
		t.Fatalf("expected params to be []any, got %T", request.Params)
	}

	if len(properties) != len(constant.StatusProperties) {
		t.Fatalf("expected %d properties, got %d", len(constant.StatusProperties), len(properties))
	}

	if properties[0] != "power" || properties[len(properties)-1] != "child_lock" {
		t.Fatalf("unexpected properties: %#v", properties)
	}
}

func TestStatusRequestsAllPropertiesAtOnce(t *testing.T) {
	t.Parallel()

	token := "00112233445566778899aabbccddeeff"
	conf, err := airconfig.New("192.168.0.10", token, time.Second)
	if err != nil {
		t.Fatalf("unexpected new config error: %v", err)
	}

	deviceID := [4]byte{0xaa, 0xbb, 0xcc, 0xdd}
	client := testclient.NewFakeClient(
		buildHelloResponse(deviceID, 200),
		buildRPCResponse(t, token, deviceID, 201, miioresponse.MIIOResponse{
			ID: 1,
			Result: mustMarshalJSON(t, []any{
				"on",
				12,
				10,
				48,
				231,
				"favorite",
				7,
				82,
				400,
				755,
				12345,
				"on",
				1,
				"off",
				"on",
			}),
		}),
	)
	airPurifierService, err := service.NewWithClient(service.Config{
		Address: conf.Address,
		Token:   conf.Token,
		Timeout: conf.Timeout,
	}, client)
	if err != nil {
		t.Fatalf("unexpected new service error: %v", err)
	}

	_, err = airPurifierService.Status(context.Background())
	if err != nil {
		t.Fatalf("unexpected status error: %v", err)
	}

	requests := client.Requests()
	if len(requests) != 2 {
		t.Fatalf("expected 2 requests, got %d", len(requests))
	}

	firstCommand := decodeRequestPayload(t, requests[1], token)
	firstParams := firstCommand.Params.([]any)

	if len(firstParams) != len(constant.StatusProperties) {
		t.Fatalf("expected a single request with %d properties, got %d", len(constant.StatusProperties), len(firstParams))
	}
}

func TestSetFavoriteLevelRejectsOutOfRange(t *testing.T) {
	t.Parallel()

	conf, err := airconfig.New("192.168.0.10", "00112233445566778899aabbccddeeff", time.Second)
	if err != nil {
		t.Fatalf("unexpected new config error: %v", err)
	}

	airPurifierService, err := service.New(service.Config{
		Address: conf.Address,
		Token:   conf.Token,
		Timeout: conf.Timeout,
	})
	if err != nil {
		t.Fatalf("unexpected new service error: %v", err)
	}

	if err := airPurifierService.SetFavoriteLevel(context.Background(), 18); !errors.Is(err, constant.ErrFavoriteLevelOutOfRange) {
		t.Fatalf("expected ErrFavoriteLevelOutOfRange, got %v", err)
	}
}

func buildHelloResponse(deviceID [4]byte, stamp uint32) []byte {
	response := make([]byte, 32)
	copy(response[0:2], []byte{0x21, 0x31})
	response[3] = 0x20
	copy(response[8:12], deviceID[:])
	response[12] = byte(stamp >> 24)
	response[13] = byte(stamp >> 16)
	response[14] = byte(stamp >> 8)
	response[15] = byte(stamp)
	return response
}

func buildRPCResponse(t *testing.T, token string, deviceID [4]byte, stamp uint32, response miioresponse.MIIOResponse) []byte {
	t.Helper()

	tokenBytes := mustToken(t, token)

	payloadBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	encryptedPayload, err := miio.Encrypt(append(payloadBytes, 0x00), tokenBytes)
	if err != nil {
		t.Fatalf("unexpected encrypt error: %v", err)
	}

	length := uint16(32 + len(encryptedPayload))
	header := make([]byte, 16)
	binary.BigEndian.PutUint16(header[0:2], 0x2131)
	binary.BigEndian.PutUint16(header[2:4], length)
	binary.BigEndian.PutUint32(header[4:8], 0)
	copy(header[8:12], deviceID[:])
	binary.BigEndian.PutUint32(header[12:16], stamp)

	checksumInput := append(append([]byte{}, header...), tokenBytes...)
	checksumInput = append(checksumInput, encryptedPayload...)
	checksum := md5Sum(checksumInput)

	responsePacket := make([]byte, 0, length)
	responsePacket = append(responsePacket, header...)
	responsePacket = append(responsePacket, checksum...)
	responsePacket = append(responsePacket, encryptedPayload...)

	return responsePacket
}

func decodeRequestPayload(t *testing.T, packetBytes []byte, token string) miiorequest.MIIORequest {
	t.Helper()

	tokenBytes := mustToken(t, token)

	parsedPacket, err := miio.ParsePacket(packetBytes)
	if err != nil {
		t.Fatalf("unexpected parse packet error: %v", err)
	}

	payloadBytes, err := miio.Decrypt(parsedPacket.Data, tokenBytes)
	if err != nil {
		t.Fatalf("unexpected decrypt error: %v", err)
	}

	request := miiorequest.MIIORequest{}
	if err := json.Unmarshal(payloadBytes, &request); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	return request
}

func mustMarshalJSON(t *testing.T, value any) json.RawMessage {
	t.Helper()

	data, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("unexpected marshal error: %v", err)
	}

	return data
}

func md5Sum(data []byte) []byte {
	sum := md5.Sum(data) //nolint:gosec
	return sum[:]
}

func mustToken(t *testing.T, token string) []byte {
	t.Helper()

	conf, err := airconfig.New("192.168.0.10", token, time.Second)
	if err != nil {
		t.Fatalf("unexpected new config error: %v", err)
	}

	return conf.Token
}
