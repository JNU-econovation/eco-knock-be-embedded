package response

type CommonResponse struct {
	IsError bool   `json:"isError"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	CommonResponse
	ErrorCode string `json:"errorCode"`
}
