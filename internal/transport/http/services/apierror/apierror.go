package apierror

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/iLeoon/realtime-gateway/pkg/logger"
)

type Code string

const (
	BadRequest          Code = "BadRequest"
	ForbiddenRequest    Code = "Forbidden"
	UnauthorizedRequest Code = "UnauthorizedRequest"
	NotFoundRequest     Code = "NotFoundRequest"
	InternalServerError Code = "InternalServerError"
	StatusNotAccepted   Code = "StatusNotAccepted"
	BadArgument         Code = "BadArgumet"
)

// APIError follows the Microsoft REST API Guidelines for error condition responses.
// See: https://github.com/microsoft/api-guidelines/blob/vNext/Guidelines.md#7102-error-condition-responses
type APIError struct {
	Error ErrorBody `json:"error"`
}

type ErrorBody struct {
	Code       string         `json:"code"`
	Message    string         `json:"message"`
	Target     string         `json:"target,omitempty"`
	Details    []ErrorDetails `json:"details,omitempty"`
	InnerError *InnerError    `json:"innererror,omitempty"`
}

type ErrorDetails struct {
	Code    string `json:"code"`
	Target  string `json:"target"`
	Message string `json:"message"`
}

type InnerError struct {
	Code       string             `json:"code"`
	InnerError *InnerErrorDetails `json:"innererror,omitempty"`
}

type InnerErrorDetails struct {
	Code   string `json:"code"`
	MinLen string `json:"minLength"`
	MaxLen string `json:"maxLength"`
}

type ErrorOptions func(*APIError)

func Send(w http.ResponseWriter, statusCode int, apiErr *APIError) {
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(apiErr); err != nil {
		logger.Error("Failed to encode the response error", "Error", err.Error())
		serverErr := APIError{
			Error: ErrorBody{
				Code:    "InternalServerError",
				Message: "Unexpected error occured while trying to process response",
				Target:  "RespErr",
				InnerError: &InnerError{
					Code: "FaildToEncodeRespError",
				},
			},
		}
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(serverErr)
		return
	}

	w.WriteHeader(statusCode)
	w.Header().Set("Content-Type", "application/json")

	_, err := w.Write(buf.Bytes())
	if err != nil {
		logger.Error("Faild to write the the response", "Error", err.Error())
	}
}

func Build(code Code, message string, opts ...ErrorOptions) *APIError {
	apiErr := &APIError{
		Error: ErrorBody{
			Code:    string(code),
			Message: message,
		},
	}

	for _, opt := range opts {
		opt(apiErr)
	}
	return apiErr
}

func WithTarget(target string) ErrorOptions {
	return func(e *APIError) {
		e.Error.Target = target
	}

}

func WithDetails(details []ErrorDetails) ErrorOptions {
	return func(e *APIError) {
		e.Error.Details = details
	}

}

func WithInnerError(innererror InnerError) ErrorOptions {
	return func(e *APIError) {
		e.Error.InnerError = &innererror
	}
}
