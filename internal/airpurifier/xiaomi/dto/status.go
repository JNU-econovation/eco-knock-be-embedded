package dto

import "eco-knock-be-embedded/internal/airpurifier/xiaomi/constant"

type StatusDTO struct {
	Power               string
	IsOn                bool
	AQI                 int
	AverageAQI          int
	Humidity            int
	Temperature         *float64
	Mode                constant.OperationMode
	FavoriteLevel       int
	FilterLifeRemaining int
	FilterHoursUsed     int
	MotorSpeed          int
	PurifyVolume        int
	LED                 bool
	LEDBrightness       *constant.LEDBrightness
	Buzzer              *bool
	ChildLock           bool
}
