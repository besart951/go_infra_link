package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string, err error)
	Fatal(msg string, err error)
	With(key string, value interface{}) Logger
}

type zerologLogger struct {
	logger zerolog.Logger
}

var _ Logger = (*zerologLogger)(nil)

func New(logLevel string) Logger {
	level := zerolog.InfoLevel
	if parsed, err := zerolog.ParseLevel(logLevel); err == nil {
		level = parsed
	}

	zerolog.TimeFieldFormat = time.RFC3339

	var output io.Writer = os.Stdout
	if os.Getenv("APP_ENV") != "production" {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	zLogger := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Logger()

	return &zerologLogger{logger: zLogger}
}

func (l *zerologLogger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *zerologLogger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *zerologLogger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *zerologLogger) Error(msg string, err error) {
	l.logger.Error().Err(err).Msg(msg)
}

func (l *zerologLogger) Fatal(msg string, err error) {
	l.logger.Fatal().Err(err).Msg(msg)
}

func (l *zerologLogger) With(key string, value interface{}) Logger {
	newLogger := l.logger.With().Interface(key, value).Logger()
	return &zerologLogger{logger: newLogger}
}
