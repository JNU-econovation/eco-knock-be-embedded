package reader

const alsResolution = 0.0672

func calculateLux(raw uint16) float64 {
	lux := float64(raw) * alsResolution
	if lux <= 1000 {
		return lux
	}

	return lux * (1.0023 + lux*0.000081488 - lux*lux*0.0000000093924 + lux*lux*lux*0.00000000000060135)
}
