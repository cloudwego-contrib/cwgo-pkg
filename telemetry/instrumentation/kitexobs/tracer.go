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
package kitexobs

import (
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"github.com/cloudwego/kitex/pkg/stats"
)

// NewServerTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewServerTracer(options ...Option) stats.Tracer {
	cfg := NewConfig(options)
	if cfg.measure == nil {
		serverDurationMeasure, err := cfg.meter.Float64Histogram(semantic.ServerDuration)
		HandleErr(err)
		labelcontrol := NewOtelLabelControl(cfg.tracer, cfg.recordSourceOperation)
		cfg.measure = metric.NewMeasure(nil, metric.NewOtelRecorder(serverDurationMeasure), labelcontrol)
	}
	return &KitexTracer{
		Measure: cfg.measure,
		cfg:     cfg,
	}
}

// NewClientTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewClientTracer(options ...Option) stats.Tracer {
	cfg := NewConfig(options)
	if cfg.measure == nil {
		clientDurationMeasure, err := cfg.meter.Float64Histogram(semantic.ClientDuration)
		HandleErr(err)
		labelcontrol := NewOtelLabelControl(cfg.tracer, cfg.recordSourceOperation)
		cfg.measure = metric.NewMeasure(nil, metric.NewOtelRecorder(clientDurationMeasure), labelcontrol)
	}
	return &KitexTracer{
		Measure: cfg.measure,
		cfg:     cfg,
	}
}
