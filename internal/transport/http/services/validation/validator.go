package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/iLeoon/realtime-gateway/internal/errors"
	"github.com/iLeoon/realtime-gateway/internal/transport/http/services/apierror"
)

type options struct {
	fieldName       string
	tagName         string
	passedParameter string
	fieldValue      interface{}
}

type codeMessage struct {
	code    string
	message string
}

var v *validator.Validate

func Init(validator *validator.Validate) {
	v = validator
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		return strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
	})
}

func Validate(data interface{}) ([]apierror.ErrorDetails, error) {
	const path errors.PathName = "validation/validator"
	const op errors.Op = "validate"
	err := v.Struct(data)
	if err == nil {
		return nil, errors.B(path, op, "failed to validate struct field", err)
	}
	errs, ok := err.(validator.ValidationErrors)
	if !ok {
		return nil, errors.B(path, op, "error passed is a validation error")

	}
	errDetails := buildErrValidationMessage(errs)
	return errDetails, nil
}

func buildErrValidationMessage(ves validator.ValidationErrors) []apierror.ErrorDetails {
	apiErr := make([]apierror.ErrorDetails, 0, len(ves))
	for _, ve := range ves {
		msg := setValidationMessageOps(&options{
			fieldName:       ve.Field(),
			tagName:         ve.Tag(),
			passedParameter: ve.Param(),
			fieldValue:      ve.Value(),
		})

		apiErr = append(apiErr, apierror.ErrorDetails{
			Code:    msg.code,
			Target:  ve.Field(),
			Message: msg.message,
		})
	}
	return apiErr
}

func setValidationMessageOps(o *options) codeMessage {
	switch o.tagName {
	case "required":
		return codeMessage{
			code:    "RequiredField",
			message: fmt.Sprintf("%s is a required field: can't be omitted or empty", o.fieldName),
		}
	case "gt":
		return codeMessage{
			code:    "InvalidValuePassed",
			message: fmt.Sprintf("%s must be greater than %v: value passed %v", o.fieldName, o.passedParameter, o.fieldValue),
		}
	case "min":
		return codeMessage{
			code:    "MinConstraint",
			message: fmt.Sprintf("%s can't be less than: %v", o.fieldName, o.passedParameter),
		}
	case "max":
		return codeMessage{
			code:    "MaxConstraint",
			message: fmt.Sprintf("%s can't be more than: %v", o.fieldName, o.passedParameter),
		}
	case "unique":
		return codeMessage{
			code:    "UniqueField",
			message: fmt.Sprintf("%s is a unique field: can't have duplicate entries: %v", o.fieldName, o.fieldValue),
		}
	case "oneof":
		return codeMessage{
			code:    "InvalidValuePassed",
			message: fmt.Sprintf("%s must be either (%v): value passed %v", o.fieldName, o.passedParameter, o.fieldValue),
		}

	default:
		return codeMessage{
			code:    "InvalidFieldData",
			message: "Invalid data was passed to the field",
		}
	}
}
