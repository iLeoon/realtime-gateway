package validation

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
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

func Validate(data interface{}) []apierror.ErrorDetails {
	err := v.Struct(data)
	if err == nil {
		return nil
	}
	errors := err.(validator.ValidationErrors)
	errDetails := buildErrValidationMessage(errors, data)
	return errDetails
}

func buildErrValidationMessage(ves validator.ValidationErrors, data interface{}) []apierror.ErrorDetails {
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
			message: fmt.Sprintf("%s is a required field: value passed %v", o.fieldName, o.fieldValue),
		}
	case "gt":
		return codeMessage{
			code:    "InvalidValuePassed",
			message: fmt.Sprintf("%s must be greater than %v: value passed %v", o.fieldName, o.passedParameter, o.fieldValue),
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
