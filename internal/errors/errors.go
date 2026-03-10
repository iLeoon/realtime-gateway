package errors

// It is inspired by the error implementation in the Upspin project.
// See: github.com/upspin/upspin

import (
	"bytes"
	"errors"
	"fmt"
)

type Kind uint8
type Op string
type PathName string

const (
	NotFound Kind = iota + 1
	TimeOut
	Internal
	Client
	Network
	ServiceUnavailable
	Forbidden
)

func (k Kind) String() string {
	switch k {
	case NotFound:
		return "no resource was found"
	case TimeOut:
		return "took so long to respond"
	case Internal:
		return "internal server error"
	case Client:
		return "client error"
	case Network:
		return "connection error"
	case ServiceUnavailable:
		return "service is down or unavailable"
	case Forbidden:
		return "forbidden"
	}
	return "unknown error"
}

func (k Kind) Error() string {
	return fmt.Sprintf("error kind: %v", int(k))
}

type Error struct {
	Path PathName
	Op   Op
	Kind Kind
	Err  error
}

func (e *Error) isZero() bool {
	return e.Path == "" && e.Op == "" && e.Kind == 0 && e.Err == nil
}
func (e *Error) Unwrap() error {
	return e.Err
}

type errorString struct {
	s string
}

// New just acs as errors.New() to pass normal strings when needed.
func New(str string) *errorString {
	return &errorString{str}
}

func (e *errorString) Error() string {
	return e.s
}

func B(args ...any) error {
	e := &Error{}

	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case PathName:
			e.Path = arg
		case Kind:
			e.Kind = arg
		case string:
			e.Err = New(arg)
		case *Error:
			copy := *arg
			e.Err = &copy
		case error:
			e.Err = arg
		}
	}

	prev, ok := e.Err.(*Error)
	if !ok {
		return e
	}

	if prev.Op == e.Op {
		prev.Op = ""
	}

	if prev.Path == e.Path {
		prev.Path = ""
	}

	if prev.Kind == e.Kind {
		prev.Kind = 0
	}

	if e.Kind == 0 {
		e.Kind = prev.Kind
		prev.Kind = 0
	}
	return e
}

func append(b *bytes.Buffer, value string) {
	b.WriteString(value)
}

func (e *Error) Error() string {
	b := new(bytes.Buffer)
	if e.Path != "" {
		b.WriteString(string(e.Path))
	}

	if e.Op != "" {
		append(b, ": ")
		b.WriteString(string(e.Op))
	}

	if e.Kind != 0 {
		append(b, ": ")
		b.WriteString(e.Kind.String())
	}

	if e.Err != nil {
		if prevErr, ok := e.Err.(*Error); ok {
			if !prevErr.isZero() {
				append(b, "\n\t")
				b.WriteString(e.Err.Error())
			}

		} else {
			append(b, ": ")
			b.WriteString(e.Err.Error())
		}
	}
	return b.String()
}

func (e *Error) Is(target error) bool {
	// Case 1: The user passed a Kind (e.g., Is(err, errors.Client))
	if targetKind, ok := target.(Kind); ok {
		return e.Kind == targetKind
	}

	// Case 2: The user passed another *Error struct
	if targetErr, ok := target.(*Error); ok {
		if targetErr.Kind != 0 && e.Kind == targetErr.Kind {
			return true
		}
	}
	// Return false to tell the standard library to keep unwrapping
	return false
}

// Is reports whether err matches the target.
// Target can be a specific error OR a generic Kind.
func Is[T any](err error, target T) bool {
	targetErr, ok := any(target).(error)
	if !ok {
		return false
	}

	return errors.Is(err, targetErr)
}

// Errorf is equivalent to fmt.Errorf, but allows clients to import only this
// package for all error handling.
func Errorf(format string, args ...interface{}) error {
	return &errorString{fmt.Sprintf(format, args...)}
}
