package middleware

import (
	"fmt"
	"runtime/debug"

	"github.com/rs/zerolog"
)

// Recovery returns a middleware that recovers from panics
func Recovery(logger *zerolog.Logger) Handler {
	return func(c ConnectionInterface) error {
		defer func() {
			if r := recover(); r != nil {
				err := fmt.Errorf("panic recovered: %v\n%s", r, debug.Stack())
				if logger != nil {
					logger.Error().Err(err).Msg("WebSocket panic recovered")
				}
				// Connection will be closed by the manager
			}
		}()
		return nil
	}
}
