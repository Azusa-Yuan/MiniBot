package log_utils

import (
	"testing"

	"github.com/rs/zerolog/log"
)

func TestXxx(t *testing.T) {
	log.Debug().Msg("This is a debug message")
	log.Info().Msg("This is an info message")
	log.Error().Msg("This is an error message")
	log.Error().Str("name", "test").Msg("")
}
