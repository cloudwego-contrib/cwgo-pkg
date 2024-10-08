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

package otelprovider

import (
	"context"
	"github.com/cloudwego/kitex/pkg/klog"
	"time"

	"github.com/cloudwego/hertz/pkg/common/hlog"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/global"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider"
	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

const (
	instrumentationNameKitex = "github.com/cloudwego-contrib/telemetry-opentelemetry/otelkitex"
	instrumentationNameHertz = "github.com/cloudwego-contrib/telemetry-opentelemetry/otelhertz"
)

var _ provider.Provider = &otelProvider{}

type otelProvider struct {
	traceExp      *otlptrace.Exporter
	metricsPusher *metric.MeterProvider
}

func (p *otelProvider) Shutdown(ctx context.Context) error {
	var err error

	if p.traceExp != nil {
		if err = p.traceExp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	if p.metricsPusher != nil {
		if err = p.metricsPusher.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	return err
}

// NewOpenTelemetryProvider Initializes an otlp trace and meter provider
func NewOpenTelemetryProvider(opts ...Option) provider.Provider {
	var (
		err           error
		traceExp      *otlptrace.Exporter
		meterProvider *metric.MeterProvider
	)

	ctx := context.TODO()

	cfg := newConfig(opts)

	if !cfg.enableTracing && !cfg.enableMetrics {
		return nil
	}

	// resource
	res := newResource(cfg)

	// propagator
	otel.SetTextMapPropagator(cfg.textMapPropagator)

	// Tracing
	if cfg.enableTracing {
		// trace client
		var traceClientOpts []otlptracegrpc.Option
		if cfg.exportEndpoint != "" {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithEndpoint(cfg.exportEndpoint))
		}
		if len(cfg.exportHeaders) > 0 {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithHeaders(cfg.exportHeaders))
		}
		if cfg.exportInsecure {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithInsecure())
		}

		traceClient := otlptracegrpc.NewClient(traceClientOpts...)

		// trace exporter
		traceExp, err = otlptrace.New(ctx, traceClient)
		if err != nil {
			hlog.Fatalf("failed to create otlp trace exporter: %s", err)
			return nil
		}

		// trace processor
		bsp := sdktrace.NewBatchSpanProcessor(traceExp)

		// trace provider
		tracerProvider := cfg.sdkTracerProvider
		if tracerProvider == nil {
			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithSampler(cfg.sampler),
				sdktrace.WithResource(res),
				sdktrace.WithSpanProcessor(bsp),
			)
		}

		otel.SetTracerProvider(tracerProvider)
	}

	// Metrics
	if cfg.enableMetrics {
		// prometheus only supports CumulativeTemporalitySelector

		var metricsClientOpts []otlpmetricgrpc.Option
		if cfg.exportEndpoint != "" {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithEndpoint(cfg.exportEndpoint))
		}
		if len(cfg.exportHeaders) > 0 {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithHeaders(cfg.exportHeaders))
		}
		if cfg.exportInsecure {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithInsecure())
		}

		meterProvider = cfg.meterProvider
		if meterProvider == nil {
			// meter exporter
			metricExp, err := otlpmetricgrpc.New(context.Background(), metricsClientOpts...)
			if cfg.enableHTTP {
				handleInitErrh(err, "Failed to create the metric exporter")
			}
			if cfg.enableRPC {
				handleInitErrk(err, "Failed to create the metric exporter")
			}
			// reader := metric.NewPeriodicReader(exporter)
			reader := metric.WithReader(metric.NewPeriodicReader(metricExp, metric.WithInterval(15*time.Second)))

			meterProvider = metric.NewMeterProvider(reader, metric.WithResource(res))
		}

		// meter pusher
		otel.SetMeterProvider(meterProvider)

		var measure cwmetric.Measure
		var metrics []cwmetric.Option
		if cfg.enableRPC {
			meter := meterProvider.Meter(
				instrumentationNameKitex,
				otelmetric.WithInstrumentationVersion(semantic.SemVersion()),
			)
			serverDurationMeasure, err := meter.Float64Histogram(semantic.BuildMetricName("rpc", cfg.instanceType, semantic.ServerDuration))
			HandleErr(err)
			serverRetryMeasure, err := meter.Float64Histogram(semantic.BuildMetricName("rpc", cfg.instanceType, semantic.ServerRetry))
			HandleErr(err)
			metrics = append(metrics,
				cwmetric.WithRecorder(semantic.RPCLatency, cwmetric.NewOtelRecorder(serverDurationMeasure)),
				cwmetric.WithRecorder(semantic.RPCRetry, cwmetric.NewOtelRecorder(serverRetryMeasure)),
			)
		}
		if cfg.enableHTTP {
			meter := meterProvider.Meter(
				instrumentationNameHertz,
				otelmetric.WithInstrumentationVersion(semantic.SemVersion()),
			)
			serverRequestCountMeasure, err := meter.Int64Counter(
				semantic.BuildMetricName("http", cfg.instanceType, semantic.RequestCount),
				otelmetric.WithUnit("count"),
				otelmetric.WithDescription("measures Incoming request count total"),
			)
			HandleErr(err)

			serverLatencyMeasure, err := meter.Float64Histogram(
				semantic.BuildMetricName("http", cfg.instanceType, semantic.ServerLatency),
				otelmetric.WithUnit("ms"),
				otelmetric.WithDescription("measures th incoming end to end duration"),
			)
			HandleErr(err)
			metrics = append(metrics,
				cwmetric.WithCounter(semantic.HTTPCounter, cwmetric.NewOtelCounter(serverRequestCountMeasure)),
				cwmetric.WithRecorder(semantic.HTTPLatency, cwmetric.NewOtelRecorder(serverLatencyMeasure)),
			)
		}

		measure = cwmetric.NewMeasure(metrics...)

		global.SetTracerMeasure(measure)

		err = runtimemetrics.Start()
		if cfg.enableHTTP {
			handleInitErrh(err, "Failed to start runtime meter collector")
		}
		if cfg.enableRPC {
			handleInitErrk(err, "Failed to start runtime meter collector")
		}

	}

	return &otelProvider{
		traceExp:      traceExp,
		metricsPusher: meterProvider,
	}
}

func newResource(cfg *config) *resource.Resource {
	if cfg.resource != nil {
		return cfg.resource
	}

	res, err := resource.New(
		context.Background(),
		resource.WithHost(),
		resource.WithFromEnv(),
		resource.WithProcessPID(),
		resource.WithTelemetrySDK(),
		resource.WithDetectors(cfg.resourceDetectors...),
		resource.WithAttributes(cfg.resourceAttributes...),
	)
	if err != nil {
		return resource.Default()
	}
	return res
}

func handleInitErrh(err error, message string) {
	if err != nil {
		hlog.Fatalf("%s: %v", message, err)
	}
}

func handleInitErrk(err error, message string) {
	if err != nil {
		klog.Fatalf("%s: %v", message, err)
	}
}

func HandleErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}
