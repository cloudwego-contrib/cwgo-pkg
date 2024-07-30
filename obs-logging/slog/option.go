package slog

import (
	cwslog "github.com/cloudwego-contrib/obs-opentelemetry/logging/slog"
	"log/slog"
)

type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	logger      *cwslog.Logger
	traceConfig *traceConfig
}

// defaultConfig default config
func defaultConfig() *config {
	return &config{
		traceConfig: &traceConfig{
			recordStackTraceInSpan: true,
			errorSpanLevel:         slog.LevelError,
		},
		logger: cwslog.NewLogger(),
	}
}

// WithLogger configures logger
func WithLogger(logger *cwslog.Logger) Option {
	return option(func(cfg *config) {
		cfg.logger = logger
	})
}

// WithTraceErrorSpanLevel trace error span level option
func WithTraceErrorSpanLevel(level slog.Level) Option {
	return option(func(cfg *config) {
		cfg.traceConfig.errorSpanLevel = level
	})
}

// WithRecordStackTraceInSpan record stack track option
func WithRecordStackTraceInSpan(recordStackTraceInSpan bool) Option {
	return option(func(cfg *config) {
		cfg.traceConfig.recordStackTraceInSpan = recordStackTraceInSpan
	})
}
