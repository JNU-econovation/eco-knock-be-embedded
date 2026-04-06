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
	ErrTimeoutRequired         = errors.New("timeout is required")
	ErrAddressRequired         = errors.New("air purifier address is required")
	ErrTokenRequired           = errors.New("air purifier token is required")
	ErrInvalidTokenLength      = errors.New("air purifier token must be 32 hex characters")
	ErrFavoriteLevelOutOfRange = errors.New("favorite level must be between 0 and 17")
	ErrUnexpectedCommandResult = errors.New("unexpected miio command result")
	HelloPacket                = miio.MustDecodeHex("21310020ffffffffffffffffffffffffffffffffffffffffffffffffffffffff")
	StatusProperties           = []string{"power", "aqi", "average_aqi", "humidity", "temp_dec", "mode", "favorite_level", "filter1_life", "f1_hour_used", "motor1_speed", "purify_volume", "led", "led_b", "buzzer", "child_lock"}
)
