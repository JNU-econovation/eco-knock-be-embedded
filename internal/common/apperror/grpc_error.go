package apperror

import (
	"errors"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/status"
)

func From(err error) *AppError {
	if err == nil {
		return nil
	}

	if appErr, ok := errors.AsType[*AppError](err); ok {
		return appErr
	}

	return New(InternalServer, err)
}

func ToGRPCError(err error) error {
	appErr := From(err)
	if appErr == nil {
		return nil
	}

	return appErr.GRPCError()
}

func (e *AppError) GRPCError() error {
	st := status.New(e.Code.GRPCCode(), e.Message)
	withDetails, err := st.WithDetails(&errdetails.ErrorInfo{
		Reason: e.CodeString(),
		Domain: string(e.Code.Domain()),
		Metadata: map[string]string{
			"errorCode": e.CodeString(),
		},
	})
	if err != nil {
		return st.Err()
	}

	return withDetails.Err()
}
