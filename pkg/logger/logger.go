package logger

import (
	"log/slog"
	"os"
)

// A global variable exposes the API surface for slog
var logger *slog.Logger

func Initlogger() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)
}

func Info(msg string, args ...any) {

	logger.Info(msg, args...)

}

func Error(msg string, args ...any) {

	logger.Error(msg, args...)

}
