package log

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

type level int

type logger struct {
	level level
}

const (
	DebugLevel level = iota
	InfoLevel
	ErrorLevel
	DisabledLevel
)

type Logger interface {
	Printf(format string, v ...any)

	Print(v ...any)

	Println(v ...any)

	Fatal(v ...any)

	Fatalf(format string, v ...any)
}

var (
	mu    sync.RWMutex
	Debug = &logger{DebugLevel}
	Info  = &logger{InfoLevel}
	Error = &logger{ErrorLevel}
	state = loggerState{currentLevel: InfoLevel, defaultLogger: newLogger(os.Stderr)}
)

type loggerState struct {
	currentLevel  level
	defaultLogger Logger
}

func lState() loggerState {
	mu.RLock()
	defer mu.RUnlock()
	return state

}

func newLogger(w io.Writer) Logger {
	return log.New(w, "", log.Ldate|log.Ltime|log.LUTC|log.Lmicroseconds)
}

type logBridge struct {
	Logger
}

func (lb logBridge) Write(b []byte) (n int, err error) {
	var message string
	parts := bytes.SplitN(b, []byte{':'}, 3)
	if len(parts) != 3 || len(parts[0]) < 1 || len(parts[2]) < 1 {
		message = fmt.Sprintf("bad log format: %s", b)
	} else {
		message = string(parts[2][1:])
	}
	lb.Print(message)
	return len(b), nil
}

func NewStdLogger(l Logger) *log.Logger {
	lb := logBridge{l}
	return log.New(lb, "", log.Lshortfile)
}

func SetSource(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()

	if w == nil {
		state.defaultLogger = nil
	} else {
		state.defaultLogger = newLogger(w)
	}
}

func (l *logger) Print(v ...any) {
	s := lState()

	if l.level < s.currentLevel {
		return
	}

	if s.defaultLogger != nil {
		s.defaultLogger.Print(v...)
	}
}

func (l *logger) Printf(format string, v ...any) {
	s := lState()

	if l.level < s.currentLevel {
		return
	}

	if s.defaultLogger != nil {
		s.defaultLogger.Printf(format, v...)
	}

}

func (l *logger) Println(v ...any) {
	s := lState()

	if l.level < s.currentLevel {
		return
	}

	if s.defaultLogger != nil {
		s.defaultLogger.Println(v...)
	}
}

func (l *logger) Fatal(v ...any) {
	s := lState()

	if s.defaultLogger != nil {
		s.defaultLogger.Fatal(v...)
	} else {
		log.Fatal(v...)
	}
}

func (l *logger) Fatalf(format string, v ...any) {
	s := lState()

	if s.defaultLogger != nil {
		s.defaultLogger.Fatalf(format, v...)
	} else {
		log.Fatalf(format, v...)
	}
}

func toString(level level) string {
	switch level {
	case InfoLevel:
		return "info"
	case DebugLevel:
		return "debug"
	case ErrorLevel:
		return "error"
	case DisabledLevel:
		return "disabled"
	}
	return "unknown"
}

func toLevel(level string) (level, error) {
	switch level {
	case "info":
		return InfoLevel, nil
	case "debug":
		return DebugLevel, nil
	case "error":
		return ErrorLevel, nil
	case "disabled":
		return DisabledLevel, nil
	}
	return DisabledLevel, fmt.Errorf("invalid log level %q", level)
}

func SetLevel(level string) error {
	l, err := toLevel(level)
	if err != nil {
		return err
	}
	mu.Lock()
	state.currentLevel = l
	mu.Unlock()
	return nil
}

// Fatal writes a message to the log and aborts.
func Fatal(v ...interface{}) {
	Info.Fatal(v...)
}

// Fatalf writes a formatted message to the log and aborts.
func Fatalf(format string, v ...interface{}) {
	Info.Fatalf(format, v...)
}
