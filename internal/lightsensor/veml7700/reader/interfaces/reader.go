package interfaces

import "eco-knock-be-embedded/internal/lightsensor/dto"

type Reader interface {
	Read() (dto.SampleDTO, error)
	Close() error
}
