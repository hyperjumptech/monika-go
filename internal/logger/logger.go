package logger

import (
	"os"

	"github.com/rs/zerolog"
)

func GetLogger() *zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return &logger
}
