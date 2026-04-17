package constant

import (
	"eco-knock-be-embedded/internal/airpurifier/xiaomi/miio"
	"errors"
)

type OperationMode string

type LEDBrightness int

type MIIOMethod string

const (
	LEDBrightnessBright LEDBrightness = 0
	LEDBrightnessDim    LEDBrightness = 1
	LEDBrightnessOff    LEDBrightness = 2

	OperationModeAuto     OperationMode = "auto"
	OperationModeSilent   OperationMode = "silent"
	OperationModeFavorite OperationMode = "favorite"
	OperationModeIdle     OperationMode = "idle"
	OperationModeMedium   OperationMode = "medium"
	OperationModeHigh     OperationMode = "high"
	OperationModeStrong   OperationMode = "strong"
	OperationModeLow      OperationMode = "low"

	MIIOMethodGetProp          MIIOMethod = "get_prop"
	MIIOMethodSetPower         MIIOMethod = "set_power"
	MIIOMethodSetMode          MIIOMethod = "set_mode"
	MIIOMethodSetLevelFavorite MIIOMethod = "set_level_favorite"
)

var (
	ErrTimeoutRequired         = errors.New("timeout 값이 필요합니다")
	ErrAddressRequired         = errors.New("공기청정기 주소가 필요합니다")
	ErrTokenRequired           = errors.New("공기청정기 토큰이 필요합니다")
	ErrInvalidTokenLength      = errors.New("공기청정기 토큰은 32자리 hex 문자열이어야 합니다")
	ErrFavoriteLevelOutOfRange = errors.New("favorite level은 0 이상 17 이하여야 합니다")
	ErrUnexpectedCommandResult = errors.New("예상하지 못한 miIO 명령 결과입니다")
	HelloPacket                = miio.MustDecodeHex("21310020ffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	StatusProperties           = []string{"power", "aqi", "average_aqi", "humidity", "temp_dec", "mode", "favorite_level", "filter1_life", "f1_hour_used", "motor1_speed", "purify_volume", "led", "led_b", "buzzer", "child_lock"}
)
