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
package prometheus

import (
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/kitexobs"
	"github.com/cloudwego/kitex/pkg/stats"
)

// NewClientTracer provide tracer for client call, addr and path is the scrape_configs for prometheus server.
func NewClientTracer(options ...Option) stats.Tracer {
	cfg := defaultConfig()
	for _, opt := range options {
		opt.apply(cfg)
	}

	return &kitexobs.KitexTracer{
		Measure: cfg.measure,
	}
}

// NewServerTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewServerTracer(options ...Option) stats.Tracer {
	cfg := defaultConfig()
	for _, opt := range options {
		opt.apply(cfg)
	}
	cfg.measure.SetLabelControl(DefaultPromLabelControl())
	return &kitexobs.KitexTracer{
		Measure: cfg.measure,
	}
}
