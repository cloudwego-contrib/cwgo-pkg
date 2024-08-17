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
	"github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/provider"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"

	"net/http"
)

var _ provider.Provider = &promProvider{}

// promProvider 结构体，包含 Prometheus 注册表和 HTTP 服务器
type promProvider struct {
	registry *prometheus.Registry
	server   *http.Server
	Measure  metric.Measure
}

// Shutdown 实现 Provider 接口的 Shutdown 方法
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

// NewPromProvider 初始化并返回一个新的 promProvider 实例
func NewPromProvider(addr string, opts ...Option) *promProvider {
	var registry *prometheus.Registry

	if registry == nil {
		registry = prometheus.NewRegistry()
	}

	cfg := newConfig(opts)
	server := &http.Server{
		Addr: addr,
	}
	if cfg.serveMux != nil {
		cfg.serveMux = http.DefaultServeMux
	}
	if !cfg.disableServer {
		cfg.serveMux.Handle(cfg.path, promhttp.HandlerFor(cfg.registry, promhttp.HandlerOpts{
			ErrorHandling: promhttp.ContinueOnError,
			Registry:      cfg.registry,
		}))
		server.Handler = cfg.serveMux
		go func() {
			if err := server.ListenAndServe(); err != nil {
				log.Fatalf("HTTP server ListenAndServe: %v", err)
				return
			}
		}()
	}
	var counter metric.Counter
	var recorder metric.Recorder
	if cfg.enableCounter {
		clientHandledCounter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: cfg.counterName,
				Help: fmt.Sprintf("Total number of requires completed by the %s, regardless of success or failure.", cfg.counterName),
			},
			[]string{semantic.LabelKeyCaller, semantic.LabelKeyCallee, semantic.LabelMethod, semantic.LabelKeyStatus, semantic.LabelKeyRetry},
		)
		cfg.registry.MustRegister(clientHandledCounter)
		counter = metric.NewPromCounter(clientHandledCounter)
	}
	if cfg.enableRecorder {
		clientHandledHistogram := prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    cfg.recorderName,
				Help:    fmt.Sprintf("Latency (microseconds) of the %s until it is finished.", cfg.recorderName),
				Buckets: cfg.buckets,
			},
			[]string{semantic.LabelKeyCaller, semantic.LabelKeyCallee, semantic.LabelMethod, semantic.LabelKeyStatus, semantic.LabelKeyRetry},
		)
		cfg.registry.MustRegister(clientHandledHistogram)
		recorder = metric.NewPromRecorder(clientHandledHistogram)
	}
	measure := metric.NewMeasure(counter, recorder, nil)
	if cfg.enableGoCollector {
		cfg.registry.MustRegister(collectors.NewGoCollector(collectors.WithGoCollectorRuntimeMetrics(cfg.runtimeMetricRules...)))
	}
	pp := &promProvider{
		registry: registry,
		server:   server,
		Measure:  measure,
	}
	return pp
}
