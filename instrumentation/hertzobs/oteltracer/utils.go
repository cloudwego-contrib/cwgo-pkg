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

package oteltracer

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol"
	"go.opentelemetry.io/otel/metric"
	oteltrace "go.opentelemetry.io/otel/trace"
	"time"
)

func ClientCounter(cfg *Config) (metric.Int64Counter, metric.Float64Histogram) {
	clientRequestCountMeasure, err := cfg.meter.Int64Counter(
		semantic.ClientRequestCount,
		metric.WithUnit("count"),
		metric.WithDescription("measures the client request count total"),
	)
	handleErr(err)

	clientLatencyMeasure, err := cfg.meter.Float64Histogram(
		semantic.ClientLatency,
		metric.WithUnit("ms"),
		metric.WithDescription("measures the duration outbound HTTP requests"),
	)
	handleErr(err)
	return clientRequestCountMeasure, clientLatencyMeasure
}

func HandleCustomResponseHandlere(cfg *Config, ctx context.Context, c *app.RequestContext) {
	if cfg.customResponseHandler != nil {
		// execute custom response handler
		cfg.customResponseHandler(ctx, c)
	}
}
func SpanFromReq(cfg *Config, ctx context.Context, req *protocol.Request) (context.Context, oteltrace.Span, time.Time) {
	start := time.Now()
	ctx, span := cfg.tracer.Start(
		ctx,
		cfg.clientSpanNameFormatter(req),
		oteltrace.WithTimestamp(start),
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)
	return ctx, span, start
}

// GetClientSpanNameFormatter ...
func GetClientSpanNameFormatter(c *Config, req *protocol.Request) string {
	return c.clientSpanNameFormatter(req)
}

func GetServerSpanNameFormatter(c *Config, rc *app.RequestContext) string {
	return c.serverSpanNameFormatter(rc)
}

func ShouldIgnore(c *Config, ctx context.Context, rc *app.RequestContext) bool {
	return c.shouldIgnore(ctx, rc)
}
