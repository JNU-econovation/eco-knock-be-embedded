package middleware

import (
	"eco-knock-be-embedded/internal/common/apperror"
	"eco-knock-be-embedded/internal/common/dto/response"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ErrorHandlerFunc func(*gin.Context) error

func WrapErrorHandler(handler ErrorHandlerFunc) gin.HandlerFunc {
	return func(context *gin.Context) {
		if err := handler(context); err != nil {
			_ = context.Error(err)
			context.Abort()
		}
	}
}

func HandleErrors() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Next()

		if context.Writer.Written() {
			return
		}

		ginErr := context.Errors.Last()
		if ginErr == nil {
			return
		}

		appErr := toAppError(ginErr.Err)
		logAppError(appErr)
		context.JSON(appErr.Status(), makeErrorResponse(appErr))
	}
}

func toAppError(err error) *apperror.AppError {
	if appErr, ok := errors.AsType[*apperror.AppError](err); ok {
		return appErr
	}

	return apperror.New(apperror.InternalServer, err)
}

func makeErrorResponse(appErr *apperror.AppError) response.ErrorResponse {
	return response.ErrorResponse{
		CommonResponse: response.CommonResponse{
			IsError: true,
			Message: appErr.Message,
		},
		ErrorCode: appErr.CodeString(),
	}
}

func logAppError(appErr *apperror.AppError) {
	if appErr.Err == nil {
		return
	}

	switch appErr.Status() {
	case http.StatusServiceUnavailable:
		log.Printf("fatal: central backend error: %v", appErr.Err)
	case http.StatusInternalServerError:
		log.Printf("fatal: internal server error: %v", appErr.Err)
	}
}
