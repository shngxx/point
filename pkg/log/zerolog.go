package log

import (
	"os"

	"github.com/getsentry/sentry-go"
	"github.com/rs/zerolog"
)

// Config holds configuration for logger
type Config struct {
	// Level sets the minimum log level (debug, info, warn, error)
	Level string `koanf:"level"`

	// SentryDSN is the Sentry DSN for error tracking (optional)
	// If empty, Sentry integration will be disabled
	SentryDSN string `koanf:"sentryDSN"`

	// SentryEnvironment sets the environment name for Sentry (e.g., "production", "development")
	SentryEnvironment string `koanf:"sentryEnvironment"`

	// SentryRelease sets the release version for Sentry
	SentryRelease string `koanf:"sentryRelease"`

	// SentrySampleRate sets the sample rate for Sentry events (0.0 to 1.0)
	// Default: 1.0 (100% of events)
	SentrySampleRate float64 `koanf:"sentrySampleRate"`

	// PrettyPrint enables pretty-printed JSON output (useful for development)
	PrettyPrint bool `koanf:"prettyPrint"`
}

// New creates a new zerolog.Logger with the given configuration and optional Sentry integration
func New(cfg Config) (*zerolog.Logger, error) {
	// Set log level
	level := zerolog.InfoLevel
	if cfg.Level != "" {
		var err error
		level, err = zerolog.ParseLevel(cfg.Level)
		if err != nil {
			level = zerolog.InfoLevel
		}
	}

	// Configure output
	var logger zerolog.Logger
	if cfg.PrettyPrint {
		output := zerolog.ConsoleWriter{Out: os.Stderr}
		logger = zerolog.New(output).With().
			Timestamp().
			Logger().
			Level(level)
	} else {
		logger = zerolog.New(os.Stderr).With().
			Timestamp().
			Logger().
			Level(level)
	}

	// Initialize Sentry if DSN is provided
	if cfg.SentryDSN != "" {
		sentryOptions := sentry.ClientOptions{
			Dsn:              cfg.SentryDSN,
			Environment:      cfg.SentryEnvironment,
			Release:          cfg.SentryRelease,
			TracesSampleRate: cfg.SentrySampleRate,
		}

		if cfg.SentrySampleRate == 0 {
			sentryOptions.TracesSampleRate = 1.0
		}

		if err := sentry.Init(sentryOptions); err != nil {
			return nil, err
		}

		// Add Sentry hook to logger
		logger = logger.Hook(&SentryHook{})
	}

	return &logger, nil
}

// MustNew creates a new zerolog.Logger with the given configuration
// It panics if initialization fails
// This is a convenience function for cases where logger initialization failure
// should cause the program to terminate immediately
func MustNew(cfg Config) *zerolog.Logger {
	logger, err := New(cfg)
	if err != nil {
		// Use standard log package for fatal error since logger failed to initialize
		// This prevents infinite recursion if logger initialization itself fails
		panic("failed to initialize logger: " + err.Error())
	}
	return logger
}

// SentryHook is a zerolog hook that sends errors to Sentry
type SentryHook struct{}

// Run implements the zerolog.Hook interface
func (h *SentryHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level >= zerolog.ErrorLevel && sentry.CurrentHub().Client() != nil {
		sentry.WithScope(func(scope *sentry.Scope) {
			// Try to extract error from event
			// Note: zerolog.Event doesn't expose fields directly in hook,
			// so we capture the message. For better error tracking, use
			// logger.Error().Err(err).Msg("message") which will be captured by hook
			sentry.CaptureMessage(msg)
		})
	}
}
