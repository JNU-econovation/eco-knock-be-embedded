package apperror

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
}

func New(code ErrorCode, err error, args ...any) *AppError {
	return &AppError{
		Code:    code,
		Message: code.Message(args...),
		Err:     err,
	}
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *AppError) Status() int {
	return e.Code.Status()
}

func (e *AppError) CodeString() string {
	return e.Code.Code()
}

func (e *AppError) Unwrap() error {
	return e.Err
}
