package interfaces

import (
	"eco-knock-be-embedded/internal/sensor/dto"
)

type Reader interface {
	Read() (dto.SampleDTO, error)
	Close() error
}
