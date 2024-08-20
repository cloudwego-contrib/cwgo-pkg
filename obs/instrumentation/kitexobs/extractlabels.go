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

	"github.com/cloudwego-contrib/cwgo-pkg/obs/instrumentation/internal"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/semantic"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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

func (o OtelLabelControl) ProcessAndInjectLabels(ctx context.Context) context.Context {
	tc := &internal.TraceCarrier{}
	tc.SetTracer(o.tracer)

	return internal.WithTraceCarrier(ctx, tc)
}

func (o OtelLabelControl) ProcessAndExtractLabels(ctx context.Context) []label.CwLabel {
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
		semantic.RPCSystemKitex,
		attribute.Key(semantic.LabelRPCMethodKey).String(ri.To().Method()),
		attribute.Key(semantic.LabelRPCCalleeKey).String(ri.To().ServiceName()),
		attribute.Key(semantic.LabelRPCCallerKey).String(ri.From().ServiceName()),

		semantic.RPCSystemKitexRecvSize.Int64(int64(st.RecvSize())),
		semantic.RPCSystemKitexSendSize.Int64(int64(st.SendSize())),
		semantic.RequestProtocolKey.String(ri.Config().TransportProtocol().String()),
	}

	// The source operation dimension maybe cause high cardinality issues
	if o.recordSourceOperation {
		attrs = append(attrs, semantic.SourceOperationKey.String(ri.From().Method()))
	}

	span.SetAttributes(attrs...)

	injectStatsEventsToSpan(span, st)

	if panicMsg, panicStack, rpcErr := parseRPCError(ri); rpcErr != nil || len(panicMsg) > 0 {
		recordErrorSpanWithStack(span, rpcErr, panicMsg, panicStack)
	}

	span.End(trace.WithTimestamp(getEndTimeOrNow(ri)))
	metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)
	return label.ToCwLabelsFromOtels(metricsAttributes)
}
