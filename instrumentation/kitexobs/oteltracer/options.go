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

package oteltracer

import (
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/kitexobs"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/kitexobs-contrib/obs-opentelemetry"
)

// Option opts for opentelemetry tracer provider
type Option interface {
	apply(cfg *Config)
}

type option func(cfg *Config)

func (fn option) apply(cfg *Config) {
	fn(cfg)
}

type Config struct {
	tracer trace.Tracer
	meter  metric.Meter

	tracerProvider    trace.TracerProvider
	meterProvider     metric.MeterProvider
	textMapPropagator propagation.TextMapPropagator

	recordSourceOperation bool
	counter               cwmetric.Counter
	recorder              cwmetric.Recorder
}

func NewConfig(opts []Option) *Config {
	cfg := DefaultConfig()

	for _, opt := range opts {
		opt.apply(cfg)
	}

	cfg.meter = cfg.meterProvider.Meter(
		instrumentationName,
		metric.WithInstrumentationVersion(kitexobs.SemVersion()),
	)

	cfg.tracer = cfg.tracerProvider.Tracer(
		instrumentationName,
		trace.WithInstrumentationVersion(kitexobs.SemVersion()),
	)

	return cfg
}

func DefaultConfig() *Config {
	return &Config{
		tracerProvider:    otel.GetTracerProvider(),
		meterProvider:     otel.GetMeterProvider(),
		textMapPropagator: otel.GetTextMapPropagator(),
	}
}
func (c Config) GetTextMapPropagator() propagation.TextMapPropagator {
	return c.textMapPropagator
}

// WithRecordSourceOperation configures record source operation dimension
func WithRecordSourceOperation(recordSourceOperation bool) Option {
	return option(func(cfg *Config) {
		cfg.recordSourceOperation = recordSourceOperation
	})
}

// WithTextMapPropagator configures propagation
func WithTextMapPropagator(p propagation.TextMapPropagator) Option {
	return option(func(cfg *Config) {
		cfg.textMapPropagator = p
	})
}