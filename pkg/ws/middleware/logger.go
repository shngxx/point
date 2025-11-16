package middleware

import (
	"github.com/rs/zerolog"
)

// Logger returns a middleware that logs WebSocket connections and messages
func Logger(l *zerolog.Logger) Handler {
	if l == nil {
		// Return no-op middleware if logger is nil
		return func(c ConnectionInterface) error {
			return nil
		}
	}

	return func(c ConnectionInterface) error {
		l.Info().Msg("WebSocket connection established")
		return nil
	}
}
