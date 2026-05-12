package apperror

import (
	"eco-knock-be-embedded/internal/common/constant"
	"fmt"
	"net/http"

	"google.golang.org/grpc/codes"
)

type ErrorCode int

const (
	InternalServer ErrorCode = iota + 1
	SensorReadFailed
	LightSensorReadFailed
	AirPurifierReadFailed
)

type Meta struct {
	Domain   constant.Domain
	Status   int
	GRPCCode codes.Code
	Number   int
	Message  string
}

var metas = map[ErrorCode]Meta{
	InternalServer: {
		Domain:   constant.DomainCommon,
		Status:   http.StatusInternalServerError,
		GRPCCode: codes.Internal,
		Number:   1,
		Message:  "내부 서버 오류가 발생했습니다",
	},
	SensorReadFailed: {
		Domain:   constant.DomainSensor,
		Status:   http.StatusServiceUnavailable,
		GRPCCode: codes.Unavailable,
		Number:   1,
		Message:  "센서 읽기에 실패했습니다",
	},
	LightSensorReadFailed: {
		Domain:   constant.DomainLightSensor,
		Status:   http.StatusServiceUnavailable,
		GRPCCode: codes.Unavailable,
		Number:   1,
		Message:  "조도 센서 읽기에 실패했습니다",
	},
	AirPurifierReadFailed: {
		Domain:   constant.DomainAirPurifier,
		Status:   http.StatusServiceUnavailable,
		GRPCCode: codes.Unavailable,
		Number:   1,
		Message:  "공기청정기 상태 조회에 실패했습니다",
	},
}

func (e ErrorCode) meta() Meta {
	return metas[e]
}

func (e ErrorCode) Code() string {
	m := e.meta()
	return fmt.Sprintf("%s_%d_%03d", m.Domain, m.Status, m.Number)
}

func (e ErrorCode) Status() int {
	return e.meta().Status
}

func (e ErrorCode) GRPCCode() codes.Code {
	return e.meta().GRPCCode
}

func (e ErrorCode) Domain() constant.Domain {
	return e.meta().Domain
}

func (e ErrorCode) Message(args ...any) string {
	m := e.meta()
	if len(args) == 0 {
		return m.Message
	}

	return fmt.Sprintf(m.Message, args...)
}
