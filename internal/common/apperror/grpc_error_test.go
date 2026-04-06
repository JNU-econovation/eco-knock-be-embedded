package apperror

import (
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestToGRPCError_AppError(t *testing.T) {
	err := ToGRPCError(New(SensorReadFailed, nil))
	if err == nil {
		t.Fatal("expected grpc error")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatal("expected grpc status error")
	}

	if st.Code() != codes.Unavailable {
		t.Fatalf("expected code %s, got %s", codes.Unavailable, st.Code())
	}

	if st.Message() != SensorReadFailed.Message() {
		t.Fatalf("expected message %q, got %q", SensorReadFailed.Message(), st.Message())
	}

	if len(st.Details()) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(st.Details()))
	}

	errorInfo, ok := st.Details()[0].(*errdetails.ErrorInfo)
	if !ok {
		t.Fatalf("expected ErrorInfo detail, got %T", st.Details()[0])
	}

	if errorInfo.Reason != SensorReadFailed.Code() {
		t.Fatalf("expected reason %s, got %s", SensorReadFailed.Code(), errorInfo.Reason)
	}

	if errorInfo.Domain != string(SensorReadFailed.Domain()) {
		t.Fatalf("expected domain %s, got %s", SensorReadFailed.Domain(), errorInfo.Domain)
	}
}
