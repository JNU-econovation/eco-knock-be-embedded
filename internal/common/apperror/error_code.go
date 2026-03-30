package apperror

import (
	"eco-knock-be-embedded/internal/common/constant"
	"fmt"
	"net/http"
)

type ErrorCode int

const (
	InternalServer ErrorCode = iota + 1
	AuthorizationHeaderRequired
	Unauthorized
	CentralBackendUnavailable
)

type Meta struct {
	Domain  constant.Domain
	Status  int
	Number  int
	Message string
}

var metas = map[ErrorCode]Meta{
	InternalServer: {
		Domain:  constant.DomainCommon,
		Status:  http.StatusInternalServerError,
		Number:  1,
		Message: "서버 내부 오류가 발생했습니다.",
	},
	AuthorizationHeaderRequired: {
		Domain:  constant.DomainCommon,
		Status:  http.StatusUnauthorized,
		Number:  2,
		Message: "Authorization 헤더가 필요합니다.",
	},
	Unauthorized: {
		Domain:  constant.DomainCommon,
		Status:  http.StatusUnauthorized,
		Number:  3,
		Message: "유효하지 않은 액세스 토큰입니다.",
	},
	CentralBackendUnavailable: {
		Domain:  constant.DomainCommon,
		Status:  http.StatusServiceUnavailable,
		Number:  1,
		Message: "중앙 백엔드를 현재 사용할 수 없습니다.",
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

func (e ErrorCode) Message(args ...any) string {
	m := e.meta()
	if len(args) == 0 {
		return m.Message
	}
	return fmt.Sprintf(m.Message, args...)
}
