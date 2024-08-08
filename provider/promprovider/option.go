package promprovider

import (
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
)

var defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}

// Option opts for opentelemetry tracer provider
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	serveMux    *http.ServeMux
	registry    *prometheus.Registry
	measure     cwmetric.Measure
	serviceName string
}

func newConfig(opts []Option) *config {
	cfg := defaultConfig()

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}

func defaultConfig() *config {
	return &config{
		registry: prometheus.NewRegistry(),
		serveMux: http.DefaultServeMux,
	}
}

// WithRegistry define your custom registry
func WithRegistry(registry *prometheus.Registry) Option {
	return option(func(cfg *config) {
		if registry != nil {
			cfg.registry = registry
		}
	})
}

// WithServeMux define your custom serve mux
func WithServeMux(serveMux *http.ServeMux) Option {
	return option(func(cfg *config) {
		if serveMux != nil {
			cfg.serveMux = serveMux
		}
	})
}

func WithServiceName(serviceName string) Option {
	return option(func(cfg *config) {
		cfg.serviceName = serviceName
	})
}

func WithMeasure(measure cwmetric.Measure) Option {
	return option(func(cfg *config) {
		cfg.measure = measure
	})
}
