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
//

package hertzobs

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/internal"
	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/adaptor"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

var _ label.LabelControl = OtelLabelControl{}

type OtelLabelControl struct {
	tracer                   trace.Tracer
	shouldIgnore             ConditionFunc
	serverHttpRouteFormatter func(c *app.RequestContext) string
}

func NewOtelLabelControl(tracer trace.Tracer, shouldIgnore ConditionFunc, serverHttpRouteFormatter func(c *app.RequestContext) string) OtelLabelControl {
	return OtelLabelControl{
		tracer:                   tracer,
		shouldIgnore:             shouldIgnore,
		serverHttpRouteFormatter: serverHttpRouteFormatter,
	}
}

func (o OtelLabelControl) ProcessAndInjectLabels(ctx context.Context) context.Context {
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if ok == false {
		return ctx
	}
	if o.shouldIgnore(ctx, c) {
		return ctx
	}
	tc := &internal.TraceCarrier{}
	tc.SetTracer(o.tracer)

	return internal.WithTraceCarrier(ctx, tc)
}

func (o OtelLabelControl) ProcessAndExtractLabels(ctx context.Context) []label.CwLabel {
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if ok == false {
		return nil
	}
	if o.shouldIgnore(ctx, c) {
		return nil
	}
	// trace carrier from context
	tc := internal.TraceCarrierFromContext(ctx)
	if tc == nil {
		logging.Debugf("get tracer container failed")
		return nil
	}

	ti := c.GetTraceInfo()
	st := ti.Stats()

	if st.Level() == stats.LevelDisabled {
		return nil
	}

	httpStart := st.GetEvent(stats.HTTPStart)
	if httpStart == nil {
		return nil
	}

	// span
	span := tc.Span()
	if span == nil || !span.IsRecording() {
		return nil
	}

	// span attributes from original http request
	if httpReq, err := adaptor.GetCompatRequest(c.GetRequest()); err == nil {
		span.SetAttributes(semconv.NetAttributesFromHTTPRequest("tcp", httpReq)...)
		span.SetAttributes(semconv.EndUserAttributesFromHTTPRequest(httpReq)...)
		span.SetAttributes(semconv.HTTPServerAttributesFromHTTPRequest("", o.serverHttpRouteFormatter(c), httpReq)...)
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
	return label.ToCwLabelsFromOtels(metricsAttributes)
}
