package mertric

import (
	"go.opentelemetry.io/otel/metric"
)

type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	// counter: <client/server>_requests_code_total{kind, operation, code, reason}
	counter metric.Int64Counter
	// histogram: <client/server>_requests_seconds_bucket{kind, operation}
	histogram metric.Float64Histogram
}

// WithTraceCounter trace error span level option
func WithTraceCounter(c metric.Int64Counter) Option {
	return option(func(cfg *config) {
		cfg.counter = c
	})
}

// WithTraceHistogram trace error span level option
func WithTraceHistogram(h metric.Float64Histogram) Option {
	return option(func(cfg *config) {
		cfg.histogram = h
	})
}

// DefaultRequestsCounter
// return metric.Int64Counter for WithRequests
// suggest histogramName = <client/server>_requests_code_total
func DefaultRequestsCounter(meter metric.Meter, histogramName string) (metric.Int64Counter, error) {
	return meter.Int64Counter(histogramName, metric.WithUnit("{call}"))
}

// DefaultSecondsHistogram
// return metric.Float64Histogram for WithSeconds
// suggest histogramName = <client/server>_requests_seconds_bucket
func DefaultSecondsHistogram(meter metric.Meter, histogramName string) (metric.Float64Histogram, error) {
	return meter.Float64Histogram(histogramName, metric.WithUnit("s"))
}
