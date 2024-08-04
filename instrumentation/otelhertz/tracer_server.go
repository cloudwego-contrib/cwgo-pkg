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

package otelhertz

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/internal"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	serverconfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/tracer"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"

	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var _ tracer.Tracer = (*serverTracer)(nil)

type serverTracer struct {
	config      *Config
	otelMetrics *cwmetric.OtelMetrics
}

func NewServerTracer(opts ...Option) (serverconfig.Option, *Config) {
	cfg := NewConfig(opts)
	st := &serverTracer{
		config: cfg,
	}

	st.createMeasures()

	return server.WithTracer(st), cfg
}

func (s *serverTracer) createMeasures() {
	serverRequestCountMeasure, err := s.config.meter.Int64Counter(
		semantic.ServerRequestCount,
		metric.WithUnit("count"),
		metric.WithDescription("measures Incoming request count total"),
	)
	handleErr(err)

	serverLatencyMeasure, err := s.config.meter.Float64Histogram(
		semantic.ServerLatency,
		metric.WithUnit("ms"),
		metric.WithDescription("measures th incoming end to end duration"),
	)
	handleErr(err)
	s.otelMetrics = cwmetric.NewOtelMetrics(serverRequestCountMeasure, serverLatencyMeasure)
}

func (s *serverTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	if s.config.shouldIgnore(ctx, c) {
		return ctx
	}
	tc := &internal.TraceCarrier{}
	tc.SetTracer(s.config.tracer)

	return internal.WithTraceCarrier(ctx, tc)
}

func (s *serverTracer) Finish(ctx context.Context, c *app.RequestContext) {
	if s.config.shouldIgnore(ctx, c) {
		return
	}
	// trace carrier from context
	tc := internal.TraceCarrierFromContext(ctx)
	if tc == nil {
		hlog.Debugf("get tracer container failed")
		return
	}

	ti := c.GetTraceInfo()
	st := ti.Stats()

	if st.Level() == stats.LevelDisabled {
		return
	}

	httpStart := st.GetEvent(stats.HTTPStart)
	if httpStart == nil {
		return
	}

	elapsedTime := float64(st.GetEvent(stats.HTTPFinish).Time().Sub(httpStart.Time())) / float64(time.Millisecond)

	// span
	span := tc.Span()
	if span == nil || !span.IsRecording() {
		return
	}

	// span attributes from original http request
	if httpReq, err := adaptor.GetCompatRequest(c.GetRequest()); err == nil {
		span.SetAttributes(semconv.NetAttributesFromHTTPRequest("tcp", httpReq)...)
		span.SetAttributes(semconv.EndUserAttributesFromHTTPRequest(httpReq)...)
		span.SetAttributes(semconv.HTTPServerAttributesFromHTTPRequest("", s.config.serverHttpRouteFormatter(c), httpReq)...)
		span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(c.Response.StatusCode()))
	}

	// span attributes
	attrs := []attribute.KeyValue{
		semconv.HTTPURLKey.String(c.URI().String()),
		semconv.NetPeerIPKey.String(c.ClientIP()),
		semconv.HTTPStatusCodeKey.Int(c.Response.StatusCode()),
	}
	span.SetAttributes(attrs...)

	injectStatsEventsToSpan(span, st)

	if panicMsg, panicStack, httpErr := parseHTTPError(ti); httpErr != nil || len(panicMsg) > 0 {
		recordErrorSpanWithStack(span, httpErr, panicMsg, panicStack)
	}

	span.End(oteltrace.WithTimestamp(getEndTimeOrNow(ti)))

	metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)
	s.otelMetrics.Inc(ctx, label.ToCwLabelsFromOtels(metricsAttributes))
	s.otelMetrics.Record(ctx, elapsedTime, label.ToCwLabelsFromOtels(metricsAttributes))
}
