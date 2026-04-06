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
		Message:  "internal server error",
	},
	SensorReadFailed: {
		Domain:   constant.DomainSensor,
		Status:   http.StatusServiceUnavailable,
		GRPCCode: codes.Unavailable,
		Number:   1,
		Message:  "sensor read failed",
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
