package errors

import "bytes"

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
	}
	return "unkown error"
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

type errorString struct {
	s string
}

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

// Is reports whether err matches the target.
// Target can be a specific error OR a generic Kind.
func Is[T any](err error, target T) bool {
	if err == nil {
		return false
	}

	switch t := any(target).(type) {
	case Kind:
		for e := err; e != nil; {
			// Check if the current error is our custom struct
			if customErr, ok := e.(*Error); ok {
				if customErr.Kind == t {
					return true
				}
				e = customErr.Err // unwrapping our struct
			} else if u, ok := e.(interface{ Unwrap() error }); ok {
				e = u.Unwrap() // unwrapping standard errors
			} else {
				break
			}
		}
		return false

	// MODE 2: Target is an ERROR (e.g., Is(err, pgx.ErrNoRows))
	case error:
		// If target is nil, we match nil errors
		if t == nil {
			return err == nil
		}

		// Look for Kind match if the target itself is a *Error
		var targetKind Kind
		if customTarget, ok := t.(*Error); ok {
			targetKind = customTarget.Kind
		}

		for e := err; e != nil; {
			// 1. Direct equality check
			if e == t {
				return true
			}

			// 2. If both are *Error, compare their Kinds
			if customErr, ok := e.(*Error); ok {
				if targetKind != 0 && customErr.Kind == targetKind {
					return true
				}
				e = customErr.Err
			} else if u, ok := e.(interface{ Unwrap() error }); ok {
				e = u.Unwrap()
			} else {
				break
			}
		}
		return false

	default:
		// If T is neither Kind nor error (e.g., a string), return false
		return false
	}
}
