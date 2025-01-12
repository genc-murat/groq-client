package groq

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrJSONEncoding   = errors.New("json encoding error")
	ErrJSONDecoding   = errors.New("json decoding error")
	ErrHTTPRequest    = errors.New("http request failed")
)

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
	Type       string `json:"type"`
}

// Error returns a formatted string representing the APIError.
// The string includes the error message, status code, and type of the error.
func (e *APIError) Error() string {
	return fmt.Sprintf("groq api error: %s (status: %d, type: %s)",
		e.Message, e.StatusCode, e.Type)
}
