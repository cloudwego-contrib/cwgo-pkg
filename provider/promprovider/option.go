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
