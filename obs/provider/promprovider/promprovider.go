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

	"github.com/cloudwego-contrib/cwgo-pkg/obs/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/provider"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/semantic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
		if cfg.enableRPC {
			RPCCounterVec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: cfg.counterName,
					Help: fmt.Sprintf("Total number of requires completed by the %s, regardless of success or failure.", cfg.counterName),
				},
				[]string{semantic.LabelRPCCallerKey, semantic.LabelRPCCalleeKey, semantic.LabelRPCMethodKey, semantic.LabelKeyStatus, semantic.LabelKeyRetry},
			)
			cfg.registry.MustRegister(RPCCounterVec)
			counter = metric.NewPromCounter(RPCCounterVec)
		} else {
			HttpCounterVec := prometheus.NewCounterVec(
				prometheus.CounterOpts{
					Name: cfg.counterName,
					Help: "Total number of HTTPs completed by the server, regardless of success or failure.",
				},
				[]string{semantic.LabelHttpMethodKey, semantic.LabelStatusCode, semantic.LabelPath},
			)
			cfg.registry.MustRegister(HttpCounterVec)
			counter = metric.NewPromCounter(HttpCounterVec)
		}
	}
	if cfg.enableRecorder {
		if cfg.enableRPC {
			clientHandledHistogramRPC := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    cfg.recorderName,
					Help:    fmt.Sprintf("Latency (microseconds) of the %s until it is finished.", cfg.recorderName),
					Buckets: cfg.buckets,
				},
				[]string{semantic.LabelRPCCallerKey, semantic.LabelRPCCalleeKey, semantic.LabelRPCMethodKey, semantic.LabelKeyStatus, semantic.LabelKeyRetry},
			)
			cfg.registry.MustRegister(clientHandledHistogramRPC)
			recorder = metric.NewPromRecorder(clientHandledHistogramRPC)
		} else {
			serverHandledHistogram := prometheus.NewHistogramVec(
				prometheus.HistogramOpts{
					Name:    cfg.recorderName,
					Help:    "Latency (microseconds) of HTTP that had been application-level handled by the server.",
					Buckets: cfg.buckets,
				},
				[]string{semantic.LabelHttpMethodKey, semantic.LabelStatusCode, semantic.LabelPath},
			)
			cfg.registry.MustRegister(serverHandledHistogram)
		}
	}
	measure := metric.NewMeasure(counter, recorder, nil)

	pp := &promProvider{
		registry: registry,
		server:   server,
		Measure:  measure,
	}
	return pp
}
