package logger_test

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"akvorado/reporter/logger"
)

func ExampleNew() {
	// Initialize zerolog
	log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.TimestampFunc = func() time.Time {
		return time.Date(2008, 1, 8, 17, 5, 05, 0, time.UTC)
	}

	// Initialize logger
	logger, err := logger.New(logger.DefaultConfiguration)
	if err != nil {
		panic(err)
	}

	logger.Info().Int("example", 15).Msg("hello world")
	// Output: {"level":"info","example":15,"time":"2008-01-08T17:05:05Z","caller":"akvorado/reporter/logger/root_example_test.go:26","module":"akvorado/reporter/logger_test","message":"hello world"}
}
