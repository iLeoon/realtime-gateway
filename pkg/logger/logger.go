package logger

import (
	"log/slog"
	"os"
)

// Package logger provides a thin wrapper around Go's slog package,
// exposing a simple global logging API for use throughout the project.

var logger *slog.Logger

func Initlogger() {
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	slog.SetDefault(logger)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}
