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

package promhertz

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/tracer"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

const (
	labelMethod       = "method"
	labelStatusCode   = "statusCode"
	labelPath         = "path"
	unknownLabelValue = "unknown"
)

// genLabels make labels values.
func genLabels(ctx *app.RequestContext) prom.Labels {
	labels := make(prom.Labels)
	labels[labelMethod] = defaultValIfEmpty(string(ctx.Request.Method()), unknownLabelValue)
	labels[labelStatusCode] = defaultValIfEmpty(strconv.Itoa(ctx.Response.Header.StatusCode()), unknownLabelValue)
	labels[labelPath] = defaultValIfEmpty(ctx.FullPath(), unknownLabelValue)

	return labels
}

func genCwLabels(ctx *app.RequestContext) []label.CwLabel {
	labels := genLabels(ctx)
	return label.ToCwLabelFromPromelabel(labels)
}

type serverTracer struct {
	promMetric cwmetric.Measure
}

// Start record the beginning of server handling request from client.
func (s *serverTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	return ctx
}

// Finish record the ending of server handling request from client.
func (s *serverTracer) Finish(ctx context.Context, c *app.RequestContext) {
	if c.GetTraceInfo().Stats().Level() == stats.LevelDisabled {
		return
	}

	httpStart := c.GetTraceInfo().Stats().GetEvent(stats.HTTPStart)
	httpFinish := c.GetTraceInfo().Stats().GetEvent(stats.HTTPFinish)
	if httpFinish == nil || httpStart == nil {
		return
	}
	cost := httpFinish.Time().Sub(httpStart.Time())
	s.promMetric.Inc(ctx, genCwLabels(c))
	s.promMetric.Record(ctx, float64(cost.Microseconds()), genCwLabels(c))
}

// NewServerTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewServerTracer(addr, path string, opts ...Option) tracer.Tracer {
	cfg := defaultConfig()

	for _, opts := range opts {
		opts.apply(cfg)
	}

	if !cfg.disableServer {
		http.Handle(path, promhttp.HandlerFor(cfg.registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
		go func() {
			if err := http.ListenAndServe(addr, nil); err != nil {
				logging.Fatal("HERTZ: Unable to start a promhttp server, err: " + err.Error())
			}
		}()
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

	measure := cwmetric.NewMeasure(cfg.counter, cfg.recorder)
	if cfg.enableGoCollector {
		cfg.registry.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(cfg.runtimeMetricRules...)))
	}

	return &serverTracer{
		promMetric: measure,
	}
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
