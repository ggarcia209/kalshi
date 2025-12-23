package kalshi

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
)

type HttpError struct {
	Code    int    `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func (e *HttpError) Error() string {
	return fmt.Sprintf("http %d: %s", e.Code, e.Message)
}

func (e *HttpError) IsClientErr() bool {
	return e.Code > 399 && e.Code < 500
}

func NewHttpError(code int, message string) *HttpError {
	return &HttpError{
		Code:    code,
		Status:  http.StatusText(code),
		Message: message,
	}
}
