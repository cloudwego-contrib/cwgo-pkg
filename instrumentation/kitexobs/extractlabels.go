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

package kitexobs

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/internal"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"go.opentelemetry.io/otel/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	labelKeyCaller = semantic.LabelKeyCaller
	labelKeyMethod = semantic.LabelMethodProm
	labelKeyCallee = semantic.LabelKeyCallee
	labelKeyStatus = semantic.LabelKeyStatus
	labelKeyRetry  = semantic.LabelKeyRetry

	// status
	statusSucceed = "succeed"
	statusError   = "error"

	unknownLabelValue = "unknown"
)

var _ label.LabelControl = OtelLabelControl{}

type OtelLabelControl struct {
	tracer                trace.Tracer
	recordSourceOperation bool
}

func NewOtelLabelControl(tracer trace.Tracer, recordoperation bool) OtelLabelControl {
	return OtelLabelControl{
		tracer:                tracer,
		recordSourceOperation: recordoperation,
	}
}

func (o OtelLabelControl) InjectLabels(ctx context.Context) context.Context {
	tc := &internal.TraceCarrier{}
	tc.SetTracer(o.tracer)

	return internal.WithTraceCarrier(ctx, tc)
}

func (o OtelLabelControl) ExtractLabels(ctx context.Context) []label.CwLabel {
	ri := rpcinfo.GetRPCInfo(ctx)
	st := ri.Stats()
	tc := internal.TraceCarrierFromContext(ctx)
	if tc == nil {
		return nil
	}
	// span
	span := tc.Span()
	if span == nil || !span.IsRecording() {
		return nil
	}

	// span attributes
	attrs := []attribute.KeyValue{
		RPCSystemKitex,
		semconv.RPCMethodKey.String(ri.To().Method()),
		semconv.RPCServiceKey.String(ri.To().ServiceName()),
		RPCSystemKitexRecvSize.Int64(int64(st.RecvSize())),
		RPCSystemKitexSendSize.Int64(int64(st.SendSize())),
		RequestProtocolKey.String(ri.Config().TransportProtocol().String()),
	}

	// The source operation dimension maybe cause high cardinality issues
	if o.recordSourceOperation {
		attrs = append(attrs, SourceOperationKey.String(ri.From().Method()))
	}

	span.SetAttributes(attrs...)

	injectStatsEventsToSpan(span, st)

	if panicMsg, panicStack, rpcErr := parseRPCError(ri); rpcErr != nil || len(panicMsg) > 0 {
		recordErrorSpanWithStack(span, rpcErr, panicMsg, panicStack)
	}

	span.End(oteltrace.WithTimestamp(getEndTimeOrNow(ri)))
	metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)
	return label.ToCwLabelsFromOtels(metricsAttributes)
}

var _ label.LabelControl = PromLabelControl{}

type PromLabelControl struct {
}

func DefaultPromLabelControl() PromLabelControl {
	return PromLabelControl{}
}

func (p PromLabelControl) InjectLabels(ctx context.Context) context.Context {
	return ctx
}

func (p PromLabelControl) ExtractLabels(ctx context.Context) []label.CwLabel {
	ri := rpcinfo.GetRPCInfo(ctx)
	extraLabels := make(prom.Labels)
	extraLabels[labelKeyStatus] = statusSucceed
	if ri.Stats().Error() != nil {
		extraLabels[labelKeyStatus] = statusError
	}
	return genCwLabels(ri)
}

func genLabels(ri rpcinfo.RPCInfo) prom.Labels {
	var (
		labels = make(prom.Labels)

		caller = ri.From()
		callee = ri.To()
	)
	labels[labelKeyCaller] = defaultValIfEmpty(caller.ServiceName(), unknownLabelValue)
	labels[labelKeyCallee] = defaultValIfEmpty(callee.ServiceName(), unknownLabelValue)
	labels[labelKeyMethod] = defaultValIfEmpty(callee.Method(), unknownLabelValue)

	labels[labelKeyStatus] = statusSucceed
	if ri.Stats().Error() != nil {
		labels[labelKeyStatus] = statusError
	}

	labels[labelKeyRetry] = "0"
	if retriedCnt, ok := callee.Tag(rpcinfo.RetryTag); ok {
		labels[labelKeyRetry] = retriedCnt
	}

	return labels
}

func genCwLabels(ri rpcinfo.RPCInfo) []label.CwLabel {
	labels := genLabels(ri)
	return label.ToCwLabelFromPromelabel(labels)
}
func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
