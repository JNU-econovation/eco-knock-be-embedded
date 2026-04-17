//go:build !linux

package client

import (
	"context"
	"crypto/md5"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/constant"
	miiorequest "eco-knock-be-embedded/internal/airpurifier/xiaomi/dto/miio/request"
	miioresponse "eco-knock-be-embedded/internal/airpurifier/xiaomi/dto/miio/response"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/miio"
	"encoding/binary"
	"encoding/json"
	"sync"
	"time"
)

type Client struct {
	config   config.Config
	mu       sync.Mutex
	deviceID [4]byte
	stamp    uint32
	state    stubState
}

type stubState struct {
	Power               string
	AQI                 int
	AverageAQI          int
	Humidity            int
	TemperatureDec      int
	Mode                string
	FavoriteLevel       int
	FilterLifeRemaining int
	FilterHoursUsed     int
	MotorSpeed          int
	PurifyVolume        int
	LED                 string
	LEDBrightness       int
	Buzzer              string
	ChildLock           string
}

func New(config config.Config) *Client {
	return &Client{
		config:   config,
		deviceID: [4]byte{0x11, 0x22, 0x33, 0x44},
		stamp:    uint32(time.Now().Unix()),
		state: stubState{
			Power:               "on",
			AQI:                 12,
			AverageAQI:          10,
			Humidity:            48,
			TemperatureDec:      231,
			Mode:                string(constant.OperationModeFavorite),
			FavoriteLevel:       7,
			FilterLifeRemaining: 82,
			FilterHoursUsed:     400,
			MotorSpeed:          755,
			PurifyVolume:        12345,
			LED:                 "on",
			LEDBrightness:       int(constant.LEDBrightnessDim),
			Buzzer:              "off",
			ChildLock:           "off",
		},
	}
}

func (client *Client) HandShake(_ context.Context, _ []byte) ([]byte, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	return client.buildHelloResponse(), nil
}

func (client *Client) Send(_ context.Context, request []byte) ([]byte, error) {
	client.mu.Lock()
	defer client.mu.Unlock()

	parsedPacket, err := miio.ParsePacket(request)
	if err != nil {
		return nil, err
	}

	payloadBytes, err := miio.Decrypt(parsedPacket.Data, client.config.Token)
	if err != nil {
		return nil, err
	}

	miioRequest := miiorequest.MIIORequest{}
	if err := json.Unmarshal(payloadBytes, &miioRequest); err != nil {
		return nil, err
	}

	response, err := client.handleRequest(miioRequest)
	if err != nil {
		return nil, err
	}

	client.stamp++
	return client.buildRPCResponse(miioRequest.ID, response), nil
}

func (client *Client) handleRequest(request miiorequest.MIIORequest) (json.RawMessage, error) {
	switch request.Method {
	case string(constant.MIIOMethodGetProp):
		return client.handleGetProperties(request.Params)
	case string(constant.MIIOMethodSetPower):
		return client.handleSetPower(request.Params)
	case string(constant.MIIOMethodSetMode):
		return client.handleSetMode(request.Params)
	case string(constant.MIIOMethodSetLevelFavorite):
		return client.handleSetFavoriteLevel(request.Params)
	default:
		return json.Marshal([]string{"ok"})
	}
}

func (client *Client) handleGetProperties(params any) (json.RawMessage, error) {
	properties, _ := params.([]any)
	results := make([]any, 0, len(properties))

	for _, property := range properties {
		results = append(results, client.propertyValue(toString(property)))
	}

	return json.Marshal(results)
}

func (client *Client) handleSetPower(params any) (json.RawMessage, error) {
	values, _ := params.([]any)
	if len(values) > 0 {
		client.state.Power = toString(values[0])
	}

	if client.state.Power != "on" {
		client.state.MotorSpeed = 0
	} else {
		client.updateMotorSpeed()
	}

	return json.Marshal([]string{"ok"})
}

func (client *Client) handleSetMode(params any) (json.RawMessage, error) {
	values, _ := params.([]any)
	if len(values) > 0 {
		client.state.Mode = toString(values[0])
	}

	client.updateMotorSpeed()
	return json.Marshal([]string{"ok"})
}

func (client *Client) handleSetFavoriteLevel(params any) (json.RawMessage, error) {
	values, _ := params.([]any)
	if len(values) > 0 {
		client.state.FavoriteLevel = toInt(values[0])
	}

	client.updateMotorSpeed()
	return json.Marshal([]string{"ok"})
}

func (client *Client) propertyValue(property string) any {
	switch property {
	case "power":
		return client.state.Power
	case "aqi":
		return client.state.AQI
	case "average_aqi":
		return client.state.AverageAQI
	case "humidity":
		return client.state.Humidity
	case "temp_dec":
		return client.state.TemperatureDec
	case "mode":
		return client.state.Mode
	case "favorite_level":
		return client.state.FavoriteLevel
	case "filter1_life":
		return client.state.FilterLifeRemaining
	case "f1_hour_used":
		return client.state.FilterHoursUsed
	case "motor1_speed":
		return client.state.MotorSpeed
	case "purify_volume":
		return client.state.PurifyVolume
	case "led":
		return client.state.LED
	case "led_b":
		return client.state.LEDBrightness
	case "buzzer":
		return client.state.Buzzer
	case "child_lock":
		return client.state.ChildLock
	default:
		return nil
	}
}

func (client *Client) updateMotorSpeed() {
	if client.state.Power != "on" {
		client.state.MotorSpeed = 0
		return
	}

	switch client.state.Mode {
	case string(constant.OperationModeFavorite):
		client.state.MotorSpeed = 400 + client.state.FavoriteLevel*50
	case string(constant.OperationModeSilent):
		client.state.MotorSpeed = 320
	case string(constant.OperationModeAuto):
		client.state.MotorSpeed = 540
	default:
		client.state.MotorSpeed = 600
	}
}

func (client *Client) buildHelloResponse() []byte {
	response := make([]byte, 32)
	binary.BigEndian.PutUint16(response[0:2], 0x2131)
	binary.BigEndian.PutUint16(response[2:4], 32)
	copy(response[8:12], client.deviceID[:])
	binary.BigEndian.PutUint32(response[12:16], client.stamp)
	return response
}

func (client *Client) buildRPCResponse(requestID int64, result json.RawMessage) []byte {
	response := miioresponse.MIIOResponse{
		ID:     requestID,
		Result: result,
	}

	payloadBytes, _ := json.Marshal(response)
	encryptedPayload, _ := miio.Encrypt(append(payloadBytes, 0x00), client.config.Token)

	packetLength := uint16(32 + len(encryptedPayload))
	header := make([]byte, 16)
	binary.BigEndian.PutUint16(header[0:2], 0x2131)
	binary.BigEndian.PutUint16(header[2:4], packetLength)
	binary.BigEndian.PutUint32(header[4:8], 0)
	copy(header[8:12], client.deviceID[:])
	binary.BigEndian.PutUint32(header[12:16], client.stamp)

	checksumInput := append(append([]byte{}, header...), client.config.Token...)
	checksumInput = append(checksumInput, encryptedPayload...)
	checksum := md5.Sum(checksumInput) //nolint:gosec

	packet := make([]byte, 0, packetLength)
	packet = append(packet, header...)
	packet = append(packet, checksum[:]...)
	packet = append(packet, encryptedPayload...)

	return packet
}

func toString(value any) string {
	if converted, ok := value.(string); ok {
		return converted
	}

	return ""
}

func toInt(value any) int {
	switch converted := value.(type) {
	case float64:
		return int(converted)
	case int:
		return converted
	default:
		return 0
	}
}
