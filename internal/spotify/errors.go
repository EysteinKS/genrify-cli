package spotify

import (
	"encoding/json"
	"fmt"
)

type apiErrorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

type APIError struct {
	Status  int
	Message string
}

func (e APIError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("spotify api error: http %d", e.Status)
	}
	return fmt.Sprintf("spotify api error: http %d: %s", e.Status, e.Message)
}

func decodeAPIError(body []byte, fallbackStatus int) error {
	var env apiErrorEnvelope
	if err := json.Unmarshal(body, &env); err == nil {
		if env.Error.Status != 0 || env.Error.Message != "" {
			status := env.Error.Status
			if status == 0 {
				status = fallbackStatus
			}
			return APIError{Status: status, Message: env.Error.Message}
		}
	}
	return APIError{Status: fallbackStatus}
}
