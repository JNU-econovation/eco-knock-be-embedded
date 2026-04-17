package miio

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"

	requestdto "eco-knock-be-embedded/internal/airpurifier/xiaomi/dto/miio/request"
)

var ErrInvalidResponse = errors.New("miIO 응답이 올바르지 않습니다")

type Packet struct {
	Length   uint16
	Unknown  uint32
	DeviceID [4]byte
	Stamp    uint32
	Checksum [16]byte
	Data     []byte
}

func MustDecodeHex(value string) []byte {
	decoded, err := hex.DecodeString(value)
	if err != nil {
		panic(err)
	}

	return decoded
}

func BuildCommandPacket(deviceID [4]byte, stamp uint32, token []byte, request requestdto.MIIORequest) ([]byte, error) {
	payloadBytes, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	encryptedPayload, err := Encrypt(append(payloadBytes, 0x00), token)
	if err != nil {
		return nil, err
	}

	packetLength := uint16(32 + len(encryptedPayload))
	header := make([]byte, 16)
	binary.BigEndian.PutUint16(header[0:2], 0x2131)
	binary.BigEndian.PutUint16(header[2:4], packetLength)
	binary.BigEndian.PutUint32(header[4:8], 0)
	copy(header[8:12], deviceID[:])
	binary.BigEndian.PutUint32(header[12:16], stamp)

	checksumInput := append(append([]byte{}, header...), token...)
	checksumInput = append(checksumInput, encryptedPayload...)
	checksum := md5.Sum(checksumInput) //nolint:gosec

	packet := make([]byte, 0, packetLength)
	packet = append(packet, header...)
	packet = append(packet, checksum[:]...)
	packet = append(packet, encryptedPayload...)

	return packet, nil
}

func ParsePacket(raw []byte) (*Packet, error) {
	if len(raw) < 32 {
		return nil, ErrInvalidResponse
	}

	if binary.BigEndian.Uint16(raw[0:2]) != 0x2131 {
		return nil, ErrInvalidResponse
	}

	var deviceID [4]byte
	copy(deviceID[:], raw[8:12])

	var checksum [16]byte
	copy(checksum[:], raw[16:32])

	return &Packet{
		Length:   binary.BigEndian.Uint16(raw[2:4]),
		Unknown:  binary.BigEndian.Uint32(raw[4:8]),
		DeviceID: deviceID,
		Stamp:    binary.BigEndian.Uint32(raw[12:16]),
		Checksum: checksum,
		Data:     append([]byte(nil), raw[32:]...),
	}, nil
}
