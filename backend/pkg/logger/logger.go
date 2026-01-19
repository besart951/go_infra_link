package logger

import (
	"log/slog"
	"os"
	"strings"
)

// env: "prod" für JSON output, "dev" für lesbaren Text
func Setup(env string, level string) *slog.Logger {
	var handler slog.Handler
	parsedLevel := parseLevel(level)

	opts := &slog.HandlerOptions{
		Level: parsedLevel,
	}

	if env == "prod" {
		// JSON für Log-Aggregatoren (Datadog, ELK, CloudWatch)
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		// Text für lokale Entwicklung (bunt und lesbar wenn unterstützt)
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	logger := slog.New(handler)

	// Optional: Setze ihn auch als globalen Default, falls mal eine 3rd Party Lib loggt
	slog.SetDefault(logger)

	return logger
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	case "info", "":
		fallthrough
	default:
		return slog.LevelInfo
	}
}
