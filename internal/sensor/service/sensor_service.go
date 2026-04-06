package service

import (
	"eco-knock-be-embedded/internal/sensor/bme680/reader/interfaces"
	"eco-knock-be-embedded/internal/sensor/dto"
)

type SensorService struct {
	reader interfaces.Reader
}

func New(reader interfaces.Reader) *SensorService {
	return &SensorService{
		reader: reader,
	}
}

func (service *SensorService) Read() dto.SensorInternalDTO {
	sample, err := service.reader.Read()
	return dto.SensorInternalDTO{
		Sample: sample,
		Err:    err,
	}
}
