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
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"

	"github.com/cloudwego/kitex/pkg/stats"
)

// Labels
const (
	labelKeyCaller = semantic.LabelKeyCaller
	labelKeyMethod = semantic.LabelMethodProm
	labelKeyCallee = semantic.LabelKeyCallee
	labelKeyStatus = semantic.LabelKeyStatus
	labelKeyRetry  = semantic.LabelKeyRetry
)

// NewClientTracer provide tracer for client call, addr and path is the scrape_configs for prometheus server.
func NewClientTracer(addr, path string, options ...Option) stats.Tracer {
	cfg := defaultConfig()
	for _, opt := range options {
		opt.apply(cfg)
	}

	if !cfg.disableServer {
		cfg.serveMux.Handle(path, promhttp.HandlerFor(cfg.registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      cfg.registry,
		}))
		go func() {
			if err := http.ListenAndServe(addr, cfg.serveMux); err != nil {
				log.Fatal("Unable to start a promhttp server, err: " + err.Error())
			}
		}()
	}

	if cfg.counter == nil {
		clientHandledCounter := prom.NewCounterVec(
			prom.CounterOpts{
				Name: semantic.ClientThroughput,
				Help: "Total number of RPCs completed by the client, regardless of success or failure.",
			},
			[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
		)
		cfg.registry.MustRegister(clientHandledCounter)
		cfg.counter = cwmetric.NewPromCounter(clientHandledCounter)
	}
	if cfg.recorder == nil {
		clientHandledHistogram := prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    semantic.ClientDuration,
				Help:    "Latency (microseconds) of the RPC until it is finished.",
				Buckets: cfg.buckets,
			},
			[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
		)
		cfg.registry.MustRegister(clientHandledHistogram)
		cfg.recorder = cwmetric.NewPromRecorder(clientHandledHistogram)
	}

	if cfg.enableGoCollector {
		cfg.registry.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(cfg.runtimeMetricRules...)))
	}
	promMetric := cwmetric.NewMeasure(cfg.counter, cfg.recorder, kitexobs.DefaultPromLabelControl())
	return &kitexobs.KitexTracer{
		Measure: promMetric,
	}
}

// NewServerTracer provides tracer for server access, addr and path is the scrape_configs for prometheus server.
func NewServerTracer(addr, path string, options ...Option) stats.Tracer {
	cfg := defaultConfig()
	for _, opt := range options {
		opt.apply(cfg)
	}

	if !cfg.disableServer {
		cfg.serveMux.Handle(path, promhttp.HandlerFor(cfg.registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      cfg.registry,
		}))
		go func() {
			if err := http.ListenAndServe(addr, cfg.serveMux); err != nil {
				log.Fatal("Unable to start a promhttp server, err: " + err.Error())
			}
		}()
	}
	if cfg.counter == nil {
		serverHandledCounter := prom.NewCounterVec(
			prom.CounterOpts{
				Name: "kitex_server_throughput",
				Help: "Total number of RPCs completed by the server, regardless of success or failure.",
			},
			[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
		)
		cfg.registry.MustRegister(serverHandledCounter)
		cfg.counter = cwmetric.NewPromCounter(serverHandledCounter)
	}
	if cfg.recorder == nil {
		serverHandledHistogram := prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "kitex_server_latency_us",
				Help:    "Latency (microseconds) of RPC that had been application-level handled by the server.",
				Buckets: cfg.buckets,
			},
			[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
		)
		cfg.registry.MustRegister(serverHandledHistogram)
		cfg.recorder = cwmetric.NewPromRecorder(serverHandledHistogram)
	}
	if cfg.enableGoCollector {
		cfg.registry.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(cfg.runtimeMetricRules...)))
	}
	measure := cwmetric.NewMeasure(cfg.counter, cfg.recorder, kitexobs.DefaultPromLabelControl())
	return &kitexobs.KitexTracer{
		Measure: measure,
	}
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
