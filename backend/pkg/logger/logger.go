package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func New(logLevel string) zerolog.Logger {
	level, err := zerolog.ParseLevel(logLevel)
	if err != nil || level == zerolog.NoLevel {
		level = zerolog.InfoLevel
	}

	zerolog.TimeFieldFormat = time.RFC3339
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

	return zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()
}
