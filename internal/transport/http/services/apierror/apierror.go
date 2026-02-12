package apierror

import (
	"net/http"

	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Code string

const (
	BadRequestCode          Code = "BadRequest"
	ForbiddenRequestCode    Code = "ForbiddenRequest"
	UnAuthorizedRequestCode Code = "UnauthorizedRequest"
	NotFoundRequestCode     Code = "NotFoundRequest"
	InternalServerErrorCode Code = "InternalServerError"
	StatusNotAcceptedCode   Code = "StatusNotAccepted"
	BadArgumentCode         Code = "BadArgumet"
	BadGatewayCode          Code = "BadGateway"
	GatewayTimeout          Code = "Timeout"
	ServiceUnavailable      Code = "ServiceUnavailable"
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
	Code       string             `json:"code"`
	InnerError *InnerErrorDetails `json:"innererror,omitempty"`
}

type ErrorOptions func(*APIError)

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

func WithInnerError(code string) ErrorOptions {
	return func(e *APIError) {
		e.Error.InnerError = &InnerError{Code: code}
	}
}

func WithInnerErrorDetails(code string, args ...string) ErrorOptions {
	return func(e *APIError) {
		e.Error.InnerError.InnerError = &InnerErrorDetails{Code: code}
	}
}

// ErrorMapper represents mapping the lower-level errors related to the database
// to the higher-level errors
func ErrorMapper(err error, target string) (*APIError, int) {
	var apiErr *APIError
	var statusCode int
	switch {
	case errors.Is(err, errors.NotFound):
		apiErr = NoDataFound(target)
		statusCode = http.StatusNotFound
	case errors.Is(err, errors.Internal):
		apiErr = UnexpectedDatabaseFailure(InternalServerErrorCode, target, "WrongSyntax")
		statusCode = http.StatusInternalServerError
	case errors.Is(err, errors.TimeOut):
		apiErr = TimeOutError(target)
		statusCode = http.StatusGatewayTimeout
	case errors.Is(err, errors.ServiceUnavailable):
		apiErr = UnexpectedDatabaseFailure(ServiceUnavailable, target, "ServiceIsUnavilableDueToManyRequests")
		statusCode = http.StatusServiceUnavailable
	case errors.Is(err, errors.Client):
		apiErr = UnexpectedDatabaseFailure(BadRequestCode, target, "UserIsPassingInvalidData")
		statusCode = http.StatusBadRequest
	case errors.Is(err, errors.Network):
		apiErr = UnexpectedDatabaseFailure(BadGatewayCode, target, "NetworkFailure")
		statusCode = http.StatusBadGateway
	}
	return apiErr, statusCode
}

func DatabaseErrorClassification(path errors.PathName, op errors.Op, err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return errors.B(path, op, errors.NotFound, err)
	}

	switch e := err.(type) {
	case *pgconn.PgError:
		switch e.Code {
		case "08006", "08001", "08000", "08003", "08007":
			return errors.B(path, op, errors.Network, err)
		case "08004", "53300":
			return errors.B(path, op, errors.ServiceUnavailable, err)
		case "57014":
			return errors.B(path, op, errors.TimeOut, err)
		case "23505", "23503":
			return errors.B(path, op, errors.Client, err)
		default:
			return errors.B(path, op, errors.Internal, err)
		}
	}
	return errors.B(path, op, errors.Internal, err)
}
