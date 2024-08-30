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

package otelkitex

import (
	"context"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/instrumentation/internal"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
)

var _ stats.Tracer = (*KitexTracer)(nil)

type KitexTracer struct {
	measure               cwmetric.Measure
	cfg                   *Config
	recordSourceOperation bool
}

// Start record the beginning of an RPC invocation.
func (s *KitexTracer) Start(ctx context.Context) context.Context {
	tc := &internal.TraceCarrier{}
	if s.cfg.tracer != nil {
		tc.SetTracer(s.cfg.tracer)
	}

	return internal.WithTraceCarrier(ctx, tc)
}

// Finish record after receiving the response of server.
func (s *KitexTracer) Finish(ctx context.Context) {
	// rpc info
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Stats().Level() == stats.LevelDisabled {
		return
	}

	st := ri.Stats()
	rpcStart := st.GetEvent(stats.RPCStart)
	rpcFinish := st.GetEvent(stats.RPCFinish)
	duration := rpcFinish.Time().Sub(rpcStart.Time())
	elapsedTime := float64(duration) / float64(time.Millisecond)

	caller := ri.From()
	callee := ri.To()
	labels := []label.CwLabel{
		{
			Key:   semantic.LabelRPCCallerKey,
			Value: defaultValIfEmpty(caller.ServiceName(), semantic.UnknownLabelValue),
		},
		{
			Key:   semantic.LabelRPCCalleeKey,
			Value: defaultValIfEmpty(callee.ServiceName(), semantic.UnknownLabelValue),
		},
		{
			Key:   semantic.LabelRPCMethodKey,
			Value: defaultValIfEmpty(callee.Method(), semantic.UnknownLabelValue),
		},
	}
	retry := label.CwLabel{
		Key:   semantic.LabelKeyRetry,
		Value: "0",
	}
	if retriedCnt, ok := callee.Tag(rpcinfo.RetryTag); ok {
		retry = label.CwLabel{
			Key:   semantic.LabelKeyRetry,
			Value: retriedCnt,
		}
	}
	labels = append(labels, retry)

	tc := internal.TraceCarrierFromContext(ctx)

	// span
	var span trace.Span
	if tc != nil {
		span = tc.Span()
		if span != nil && span.IsRecording() {
			// span attributes
			attrs := []attribute.KeyValue{
				semantic.RPCSystemKitex,
				semantic.RPCSystemKitexRecvSize.Int64(int64(st.RecvSize())),
				semantic.RPCSystemKitexSendSize.Int64(int64(st.SendSize())),
				semantic.RequestProtocolKey.String(ri.Config().TransportProtocol().String()),
			}

			// The source operation dimension maybe cause high cardinality issues
			if s.recordSourceOperation {
				attrs = append(attrs, semantic.SourceOperationKey.String(ri.From().Method()))
			}

			span.SetAttributes(attrs...)

			injectStatsEventsToSpan(span, st)

			if panicMsg, panicStack, rpcErr := parseRPCError(ri); rpcErr != nil || len(panicMsg) > 0 {
				recordErrorSpanWithStack(span, rpcErr, panicMsg, panicStack)
			}

			span.End(trace.WithTimestamp(getEndTimeOrNow(ri)))
			metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)
			spanlabels := label.ToCwLabelsFromOtels(metricsAttributes)

			labels = append(labels, spanlabels...)
		}
	}
	if span == nil || !span.IsRecording() {
		stateless := label.CwLabel{
			Key:   semantic.LabelKeyStatus,
			Value: semantic.StatusSucceed,
		}
		if ri.Stats().Error() != nil {
			stateless = label.CwLabel{
				Key:   semantic.LabelKeyStatus,
				Value: semantic.StatusError,
			}
		}
		labels = append(labels, stateless)
	}

	// measure
	s.measure.Inc(ctx, labels)
	s.measure.Record(ctx, elapsedTime, labels)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
