package utils

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	Info  zerolog.Logger
	Warn  zerolog.Logger
	Error zerolog.Logger
)

func InitLogger() {
	base := zerolog.New(os.Stdout).With().Timestamp().Logger()

	Info = base.Level(zerolog.InfoLevel).With().Str("level", "INFO").Logger()
	Warn = base.Level(zerolog.WarnLevel).With().Str("level", "WARN").Logger()
	Error = base.Level(zerolog.ErrorLevel).With().Str("level", "ERROR").Logger()

	zerolog.TimeFieldFormat = time.RFC3339
}
