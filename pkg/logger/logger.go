package logger

import (
	"log/slog"
	"os"
)

type Logger struct{ *slog.Logger }

func New(level string) *Logger {
	var l slog.Level
	switch level {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}

	return &Logger{slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l}))}
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.Error(msg, args...)
	os.Exit(1)
}
