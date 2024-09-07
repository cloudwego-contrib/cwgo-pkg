/*
 * Copyright 2021 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package prometheus provides the extend implement of prometheus.
package otelkitex

import (
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
)

// NewServerTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewServerTracer(options ...Option) *KitexTracer {
	cfg := NewConfig(options)
	if cfg.measure == nil {
		serverDurationMeasure, err := cfg.meter.Float64Histogram(semantic.ServerDuration)
		HandleErr(err)
		serverRetryMeasure, err := cfg.meter.Float64Histogram(semantic.ServerRetry)
		HandleErr(err)
		cfg.measure = metric.NewMeasure(
			metric.WithRecorder(semantic.Latency, metric.NewOtelRecorder(serverDurationMeasure)),
			metric.WithRecorder(semantic.Retry, metric.NewOtelRecorder(serverRetryMeasure)),
		)
	}
	return &KitexTracer{
		measure: cfg.measure,
		cfg:     cfg,
	}
}

// NewClientTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewClientTracer(options ...Option) *KitexTracer {
	cfg := NewConfig(options)
	if cfg.measure == nil {
		clientDurationMeasure, err := cfg.meter.Float64Histogram(semantic.ClientDuration)
		HandleErr(err)
		clientRetryMeasure, err := cfg.meter.Float64Histogram(semantic.ClientRetry)
		HandleErr(err)
		cfg.measure = metric.NewMeasure(
			metric.WithRecorder(semantic.Latency, metric.NewOtelRecorder(clientDurationMeasure)),
			metric.WithRecorder(semantic.Retry, metric.NewOtelRecorder(clientRetryMeasure)),
		)
	}
	return &KitexTracer{
		measure: cfg.measure,
		cfg:     cfg,
	}
}
