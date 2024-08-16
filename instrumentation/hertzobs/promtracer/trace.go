/*
 * Copyright 2022 CloudWeGo Authors
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

package promtracer

import (
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/hertzobs"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/cloudwego/hertz/pkg/common/tracer"
	prom "github.com/prometheus/client_golang/prometheus"
)

const (
	labelMethod       = "method"
	labelStatusCode   = "statusCode"
	labelPath         = "path"
	unknownLabelValue = "unknown"
)

// NewServerTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewServerTracer(opts ...Option) tracer.Tracer {
	cfg := defaultConfig()

	for _, opts := range opts {
		opts.apply(cfg)
	}

	if cfg.counter == nil {
		serverHandledCounter := prom.NewCounterVec(
			prom.CounterOpts{
				Name: semantic.ServerRequestCount,
				Help: "Total number of HTTPs completed by the server, regardless of success or failure.",
			},
			[]string{labelMethod, labelStatusCode, labelPath},
		)
		cfg.registry.MustRegister(serverHandledCounter)
		cfg.counter = cwmetric.NewPromCounter(serverHandledCounter)
	}

	if cfg.counter == nil {
		serverHandledHistogram := prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    semantic.ServerLatency,
				Help:    "Latency (microseconds) of HTTP that had been application-level handled by the server.",
				Buckets: cfg.buckets,
			},
			[]string{labelMethod, labelStatusCode, labelPath},
		)
		cfg.registry.MustRegister(serverHandledHistogram)
		cfg.recorder = cwmetric.NewPromRecorder(serverHandledHistogram)
	}
	labelControl := hertzobs.DefaultPromLabelControl()
	measure := cwmetric.NewMeasure(cfg.counter, cfg.recorder, labelControl)

	return &hertzobs.HertzTracer{
		Measure: measure,
	}
}
