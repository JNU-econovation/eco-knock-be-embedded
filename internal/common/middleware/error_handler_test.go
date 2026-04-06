package middleware

import (
	"eco-knock-be-embedded/internal/common/apperror"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHandleErrors_AppError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(HandleErrors())
	router.GET("/test", WrapErrorHandler(func(context *gin.Context) error {
		return apperror.New(apperror.SensorReadFailed, nil)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, recorder.Code)
	}

	var responseBody struct {
		IsError   bool   `json:"isError"`
		Message   string `json:"message"`
		ErrorCode string `json:"errorCode"`
	}

	if err := json.Unmarshal(recorder.Body.Bytes(), &responseBody); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !responseBody.IsError {
		t.Fatalf("expected isError=true")
	}

	if responseBody.ErrorCode != apperror.SensorReadFailed.Code() {
		t.Fatalf("expected error code %s, got %s", apperror.SensorReadFailed.Code(), responseBody.ErrorCode)
	}
}

func TestHandleErrors_UnknownError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(HandleErrors())
	router.GET("/test", WrapErrorHandler(func(context *gin.Context) error {
		return errors.New("boom")
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)

	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, recorder.Code)
	}
}
