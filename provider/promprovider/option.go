// Copyright 2022 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promprovider

import (
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
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
	buckets  []float64
	serveMux *http.ServeMux
	registry *prometheus.Registry
	measure  cwmetric.Measure

	enableCounter  bool
	counterName    string
	enableRecorder bool
	recorderName   string

	disableServer      bool
	path               string
	enableGoCollector  bool
	runtimeMetricRules []collectors.GoRuntimeMetricsRule
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
		buckets:            defaultBuckets,
		registry:           prometheus.NewRegistry(),
		serveMux:           http.DefaultServeMux,
		runtimeMetricRules: []collectors.GoRuntimeMetricsRule{},
		disableServer:      false,
		counterName:        "counter",
		recorderName:       "recorder",
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

func WithCounter() Option {
	return option(func(cfg *config) {
		cfg.enableCounter = true
	})
}

func WithCounterName(name string) Option {
	return option(func(cfg *config) {
		cfg.counterName = name
	})
}

func WithRecorder() Option {
	return option(func(cfg *config) {
		cfg.enableRecorder = true
	})
}

func WithRecorderName(name string) Option {
	return option(func(cfg *config) {
		cfg.recorderName = name
	})
}

func WithMeasure(measure cwmetric.Measure) Option {
	return option(func(cfg *config) {
		cfg.measure = measure
	})
}

// WithDisableServer disable prometheus server
func WithDisableServer(disable bool) Option {
	return option(func(cfg *config) {
		cfg.disableServer = disable
	})
}

func WithPath(path string) Option {
	return option(func(cfg *config) {
		cfg.path = path
	})
}

// WithGoCollectorRule define your custom go collector rule
func WithGoCollectorRule(rules ...collectors.GoRuntimeMetricsRule) Option {
	return option(func(cfg *config) {
		cfg.runtimeMetricRules = rules
	})
}
