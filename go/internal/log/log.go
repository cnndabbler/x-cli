package log

import (
	"log/slog"
	"os"
	"strings"
)

// Init sets up slog based on X_CLI_LOG_LEVEL env var.
func Init() {
	level := slog.LevelWarn
	switch strings.ToLower(os.Getenv("X_CLI_LOG_LEVEL")) {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn", "warning":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}
