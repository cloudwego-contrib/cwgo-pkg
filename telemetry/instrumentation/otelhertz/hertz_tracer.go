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
	"strconv"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/instrumentation/internal"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/cloudwego/hertz/pkg/common/tracer"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
)

var _ tracer.Tracer = (*HertzTracer)(nil)

const requestContextKey = "requestContext"

type HertzTracer struct {
	measure cwmetric.Measure
	cfg     *Config
}

func (h HertzTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	if h.cfg.shouldIgnore(ctx, c) {
		return ctx
	}
	tc := &internal.TraceCarrier{}
	tc.SetTracer(h.cfg.tracer)

	return internal.WithTraceCarrier(ctx, tc)
}

func (h HertzTracer) Finish(ctx context.Context, c *app.RequestContext) {
	ctx = context.WithValue(ctx, requestContextKey, c)
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
	labels := []label.CwLabel{
		{
			Key:   semantic.LabelStatusCode,
			Value: defaultValIfEmpty(strconv.Itoa(c.Response.Header.StatusCode()), semantic.UnknownLabelValue),
		},
		{
			Key:   semantic.LabelPath,
			Value: defaultValIfEmpty(c.FullPath(), semantic.UnknownLabelValue),
		},
		{
			Key:   semantic.LabelHttpMethodKey,
			Value: defaultValIfEmpty(string(c.Request.Method()), semantic.UnknownLabelValue),
		},
	}

	if h.cfg.labelFunc != nil {
		labels = append(labels, h.cfg.labelFunc(c)...)
	}

	tc := internal.TraceCarrierFromContext(ctx)
	var span trace.Span
	if tc != nil && tc.Span() != nil && tc.Span().IsRecording() {
		span = tc.Span()
		// span attributes from original http request
		if httpReq, err := adaptor.GetCompatRequest(c.GetRequest()); err == nil {
			span.SetAttributes(semconv.NetAttributesFromHTTPRequest("tcp", httpReq)...)
			span.SetAttributes(semconv.EndUserAttributesFromHTTPRequest(httpReq)...)
			span.SetAttributes(semconv.HTTPServerAttributesFromHTTPRequest("", h.cfg.serverHttpRouteFormatter(c), httpReq)...)
		}

		// span attributes
		attrs := []attribute.KeyValue{
			semconv.HTTPURLKey.String(c.URI().String()),
			semconv.NetPeerIPKey.String(c.ClientIP()),
		}
		span.SetAttributes(attrs...)

		injectStatsEventsToSpan(span, st)

		if panicMsg, panicStack, httpErr := parseHTTPError(ti); httpErr != nil || len(panicMsg) > 0 {
			recordErrorSpanWithStack(span, httpErr, panicMsg, panicStack)
		}

		span.End(trace.WithTimestamp(getEndTimeOrNow(ti)))

		metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)

		labels = append(labels, label.ToCwLabelsFromOtels(metricsAttributes)...)
	}
	h.measure.Inc(ctx, semantic.Counter, labels...)
	h.measure.Record(ctx, semantic.Latency, elapsedTime, labels...)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
