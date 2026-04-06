package service

import (
	"context"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/client"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/constant"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	configdto "eco-knock-be-embedded/internal/airpurifier/xiaomi/config"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/dto"
	miiorequest "eco-knock-be-embedded/internal/airpurifier/xiaomi/dto/miio/request"
	miioresponse "eco-knock-be-embedded/internal/airpurifier/xiaomi/dto/miio/response"
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/miio"
)

type Config = configdto.Config
type DiscoveryDTO = dto.DiscoveryDTO
type StatusDTO = dto.StatusDTO
type DeviceError = miioresponse.DeviceError

type XiaomiAirPurifierService struct {
	client    client.Requester
	token     []byte
	requestID int64
	sessionMu sync.Mutex
	session   session
}

type session struct {
	discovered bool
	deviceID   [4]byte
	stamp      uint32
}

func New(config Config) (*XiaomiAirPurifierService, error) {
	return NewWithClient(config, nil)
}

func NewWithClient(
	config Config,
	miioClient client.Requester,
) (*XiaomiAirPurifierService, error) {
	if config.Address == "" {
		return nil, constant.ErrAddressRequired
	}

	if config.Token == nil {
		return nil, constant.ErrTokenRequired
	}

	timeout := config.Timeout
	if timeout <= 0 {
		return nil, constant.ErrTimeoutRequired
	}

	if miioClient == nil {
		miioClient = client.New(config)
	}

	return &XiaomiAirPurifierService{
		client:    miioClient,
		token:     config.Token,
		requestID: time.Now().Unix(),
	}, nil
}

func (service *XiaomiAirPurifierService) HandShake(ctx context.Context) (DiscoveryDTO, error) {
	response, err := service.client.HandShake(ctx, constant.HelloPacket)
	if err != nil {
		return DiscoveryDTO{}, err
	}

	parsedPacket, err := miio.ParsePacket(response)
	if err != nil {
		return DiscoveryDTO{}, err
	}

	service.sessionMu.Lock()
	service.session = session{
		discovered: true,
		deviceID:   parsedPacket.DeviceID,
		stamp:      parsedPacket.Stamp,
	}
	service.sessionMu.Unlock()

	return DiscoveryDTO{
		DeviceID: hex.EncodeToString(parsedPacket.DeviceID[:]),
		Stamp:    parsedPacket.Stamp,
	}, nil
}

func (service *XiaomiAirPurifierService) Status(ctx context.Context) (StatusDTO, error) {
	response, err := service.sendCommand(ctx, constant.MIIOMethodGetProp, constant.StatusProperties)
	if err != nil {
		return StatusDTO{}, err
	}

	var values []any
	if err := json.Unmarshal(response.Result, &values); err != nil {
		return StatusDTO{}, err
	}

	status := StatusDTO{
		Power:               stringValue(values[0]),
		IsOn:                stringValue(values[0]) == "on",
		AQI:                 intValue(values[1]),
		AverageAQI:          intValue(values[2]),
		Humidity:            intValue(values[3]),
		Mode:                constant.OperationMode(stringValue(values[5])),
		FavoriteLevel:       intValue(values[6]),
		FilterLifeRemaining: intValue(values[7]),
		FilterHoursUsed:     intValue(values[8]),
		MotorSpeed:          intValue(values[9]),
		PurifyVolume:        intValue(values[10]),
		LED:                 stringValue(values[11]) == "on",
		ChildLock:           stringValue(values[14]) == "on",
	}

	if temp := nullableTemperature(values[4]); temp != nil {
		status.Temperature = temp
	}

	if ledBrightness := nullableLEDBrightness(values[12]); ledBrightness != nil {
		status.LEDBrightness = ledBrightness
	}

	if buzzer := nullableOnOff(values[13]); buzzer != nil {
		status.Buzzer = buzzer
	}

	return status, nil
}

func (service *XiaomiAirPurifierService) SetPower(ctx context.Context, on bool) error {
	value := "off"
	if on {
		value = "on"
	}

	return service.expectOK(ctx, constant.MIIOMethodSetPower, []string{value})
}

func (service *XiaomiAirPurifierService) SetMode(ctx context.Context, mode constant.OperationMode) error {
	return service.expectOK(ctx, constant.MIIOMethodSetMode, []string{string(mode)})
}

func (service *XiaomiAirPurifierService) SetFavoriteLevel(ctx context.Context, level int) error {
	if level < 0 || level > 17 {
		return constant.ErrFavoriteLevelOutOfRange
	}

	return service.expectOK(ctx, constant.MIIOMethodSetLevelFavorite, []int{level})
}

func (service *XiaomiAirPurifierService) expectOK(ctx context.Context, method constant.MIIOMethod, params any) error {
	response, err := service.sendCommand(ctx, method, params)
	if err != nil {
		return err
	}

	var result []string
	if err := json.Unmarshal(response.Result, &result); err != nil {
		return err
	}

	if len(result) != 1 || result[0] != "ok" {
		return constant.ErrUnexpectedCommandResult
	}

	return nil
}

func (service *XiaomiAirPurifierService) sendCommand(ctx context.Context, method constant.MIIOMethod, params any) (*miioresponse.MIIOResponse, error) {
	currentSession, err := service.ensureSession(ctx)
	if err != nil {
		return nil, err
	}

	request := miiorequest.MIIORequest{
		ID:     service.nextRequestID(),
		Method: string(method),
		Params: params,
	}

	packetBytes, err := miio.BuildCommandPacket(
		currentSession.deviceID,
		currentSession.stamp+1,
		service.token,
		request,
	)
	if err != nil {
		return nil, err
	}

	responseBytes, err := service.client.Send(ctx, packetBytes)
	if err != nil {
		service.clearSession()
		return nil, err
	}

	parsedPacket, err := miio.ParsePacket(responseBytes)
	if err != nil {
		return nil, err
	}

	payloadBytes, err := miio.Decrypt(parsedPacket.Data, service.token)
	if err != nil {
		return nil, err
	}

	response := miioresponse.MIIOResponse{}
	if err := json.Unmarshal(payloadBytes, &response); err != nil {
		return nil, err
	}

	if response.Error != nil {
		return nil, response.Error
	}

	service.sessionMu.Lock()
	service.session.stamp = parsedPacket.Stamp
	service.sessionMu.Unlock()

	return &response, nil
}

func (service *XiaomiAirPurifierService) ensureSession(ctx context.Context) (session, error) {
	service.sessionMu.Lock()
	if service.session.discovered {
		currentSession := service.session
		service.sessionMu.Unlock()
		return currentSession, nil
	}
	service.sessionMu.Unlock()

	if _, err := service.HandShake(ctx); err != nil {
		return session{}, err
	}

	service.sessionMu.Lock()
	currentSession := service.session
	service.sessionMu.Unlock()

	return currentSession, nil
}

func (service *XiaomiAirPurifierService) clearSession() {
	service.sessionMu.Lock()
	service.session = session{}
	service.sessionMu.Unlock()
}

func (service *XiaomiAirPurifierService) nextRequestID() int64 {
	service.sessionMu.Lock()
	defer service.sessionMu.Unlock()

	service.requestID++
	if service.requestID >= 9999 {
		service.requestID = 1
	}

	return service.requestID
}

func stringValue(value any) string {
	if value == nil {
		return ""
	}

	switch converted := value.(type) {
	case string:
		return converted
	default:
		return fmt.Sprintf("%v", converted)
	}
}

func intValue(value any) int {
	switch converted := value.(type) {
	case nil:
		return 0
	case float64:
		return int(converted)
	case int:
		return converted
	case string:
		parsed, err := strconv.Atoi(converted)
		if err == nil {
			return parsed
		}
		return 0
	default:
		return 0
	}
}

func nullableTemperature(value any) *float64 {
	switch converted := value.(type) {
	case nil:
		return nil
	case float64:
		temperature := converted / 10.0
		return &temperature
	default:
		return nil
	}
}

func nullableLEDBrightness(value any) *constant.LEDBrightness {
	switch converted := value.(type) {
	case nil:
		return nil
	case float64:
		return new(constant.LEDBrightness(int(converted)))
	default:
		return nil
	}
}

func nullableOnOff(value any) *bool {
	switch converted := value.(type) {
	case nil:
		return nil
	case string:
		switch converted {
		case "on":
			return new(true)
		case "off":
			return new(false)
		default:
			return nil
		}
	default:
		return nil
	}
}
