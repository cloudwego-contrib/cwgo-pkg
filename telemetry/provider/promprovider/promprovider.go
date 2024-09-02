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
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var _ provider.Provider = &promProvider{}

// promProvider Structure of promProvider, including Prometheus registry and HTTP server
type promProvider struct {
	registry *prometheus.Registry
	server   *http.Server
	Measure  metric.Measure
}

// Shutdown Implement the Shutdown method for the Provider interface
func (p *promProvider) Shutdown(ctx context.Context) error {
	// 关闭 HTTP 服务器
	if err := p.server.Shutdown(ctx); err != nil {
		return err
	}

	return nil
}

func (p *promProvider) GetRegistry() *prometheus.Registry {
	return p.registry
}

// NewPromProvider Initialize and return a new promProvider instance
func NewPromProvider(addr string, opts ...Option) *promProvider {
	cfg := newConfig(opts)
	registry := cfg.registry
	if registry == nil {
		registry = prometheus.NewRegistry()
	}
	server := &http.Server{
		Addr: addr,
	}
	cfg.mu.Lock()
	if cfg.serveMux != nil {
		cfg.serveMux = http.DefaultServeMux
	}
	if !cfg.disableServer {
		cfg.serveMux.Handle(cfg.path, promhttp.HandlerFor(registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      registry,
		}))
		server.Handler = cfg.serveMux
		go func() {
			if err := server.ListenAndServe(); err != nil {
				log.Fatalf("HTTP server ListenAndServe: %v", err)
				return
			}
		}()
	}
	cfg.mu.Unlock()
	var counter metric.Counter
	var recorder metric.Recorder
	var retryRecorder metric.RetryRecorder
	var measure metric.Measure
	if cfg.enableRPC {
		if cfg.enableCounter {
			RPCCounterVec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: cfg.counterName,
					Help: fmt.Sprintf("Total number of requires completed by the %s, regardless of success or failure.", cfg.counterName),
				},
				[]string{semantic.LabelRPCCallerKey, semantic.LabelRPCCalleeKey, semantic.LabelRPCMethodKey, semantic.LabelKeyStatus},
			)
			registry.MustRegister(RPCCounterVec)
			counter = metric.NewPromCounter(RPCCounterVec)
		}
		if cfg.enableRecorder {
			clientHandledHistogramRPC := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    cfg.recorderName,
					Help:    fmt.Sprintf("Latency (microseconds) of the %s until it is finished.", cfg.recorderName),
					Buckets: cfg.buckets,
				},
				[]string{semantic.LabelRPCCallerKey, semantic.LabelRPCCalleeKey, semantic.LabelRPCMethodKey, semantic.LabelKeyStatus},
			)
			registry.MustRegister(clientHandledHistogramRPC)
			recorder = metric.NewPromRecorder(clientHandledHistogramRPC)
			// create retry recorder
			retryHandledHistogramRPC := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    fmt.Sprintf("%s_retry_attempts", cfg.recorderName),
					Help:    fmt.Sprintf("Distribution of retry attempts for %s until it is finished.", cfg.recorderName),
					Buckets: retryBuckets,
				},
				[]string{semantic.LabelRPCCallerKey, semantic.LabelRPCCalleeKey, semantic.LabelRPCMethodKey},
			)
			registry.MustRegister(clientHandledHistogramRPC)
			retryRecorder = metric.NewPromRetryRecorder(retryHandledHistogramRPC)
		}
		measure = metric.NewMeasure(counter, recorder, retryRecorder)
	} else {
		if cfg.enableCounter {
			HttpCounterVec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: cfg.counterName,
					Help: "Total number of HTTPs completed by the server, regardless of success or failure.",
				},
				[]string{semantic.LabelHttpMethodKey, semantic.LabelStatusCode, semantic.LabelPath},
			)
			registry.MustRegister(HttpCounterVec)
			counter = metric.NewPromCounter(HttpCounterVec)
		}
		if cfg.enableRecorder {
			HttpHandledHistogram := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    cfg.recorderName,
					Help:    "Latency (microseconds) of HTTP that had been application-level handled by the server.",
					Buckets: cfg.buckets,
				},
				[]string{semantic.LabelHttpMethodKey, semantic.LabelStatusCode, semantic.LabelPath},
			)
			registry.MustRegister(HttpHandledHistogram)
			recorder = metric.NewPromRecorder(HttpHandledHistogram)
		}
		measure = metric.NewMeasure(counter, recorder, nil)
	}

	pp := &promProvider{
		registry: registry,
		server:   server,
		Measure:  measure,
	}
	return pp
}
