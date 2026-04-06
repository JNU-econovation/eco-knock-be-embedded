package response

import (
	"encoding/json"
	"fmt"
)

type MIIOResponse struct {
	ID     int64           `json:"id"`
	Result json.RawMessage `json:"result"`
	Error  *DeviceError    `json:"error"`
}

type DeviceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err *DeviceError) Error() string {
	return fmt.Sprintf("miio device error %d: %s", err.Code, err.Message)
}
