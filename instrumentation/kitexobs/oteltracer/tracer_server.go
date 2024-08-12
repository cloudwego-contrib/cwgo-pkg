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
	"github.com/cloudwego-contrib/cwgo-pkg/instrumentation/kitexobs"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/cloudwego/kitex/server"
)

/*var _ stats.Tracer = (*serverTracer)(nil)

type serverTracer struct {
	Measure cwmetric.Measure
}*/

func NewServerOption(opts ...Option) (server.Option, *Config) {
	cfg := NewConfig(opts)
	st := &kitexobs.KitexTracer{}
	serverDurationMeasure, err := cfg.meter.Float64Histogram(semantic.ServerDuration)
	kitexobs.HandleErr(err)
	labelcontrol := kitexobs.NewOtelLabelControl(cfg.tracer, cfg.recordSourceOperation)
	st.Measure = cwmetric.NewMeasure(nil, cwmetric.NewOtelRecorder(serverDurationMeasure), labelcontrol)

	return server.WithTracer(st), cfg
}

/*func (s *serverTracer) Start(ctx context.Context) context.Context {
	return s.Measure.InjectLabels(ctx)
}

func (s *serverTracer) Finish(ctx context.Context) {
	// trace carrier from context
	tc := internal.TraceCarrierFromContext(ctx)
	if tc == nil {
		return
	}

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

	// span
	span := tc.Span()
	if span == nil || !span.IsRecording() {
		return
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
	if s.config.recordSourceOperation {
		attrs = append(attrs, SourceOperationKey.String(ri.From().Method()))
	}

	span.SetAttributes(attrs...)

	injectStatsEventsToSpan(span, st)

	if panicMsg, panicStack, rpcErr := parseRPCError(ri); rpcErr != nil || len(panicMsg) > 0 {
		recordErrorSpanWithStack(span, rpcErr, panicMsg, panicStack)
	}

	span.End(oteltrace.WithTimestamp(getEndTimeOrNow(ri)))

	metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)
	labels := s.Measure.ExtractLabels(ctx)
	s.Measure.Record(ctx, elapsedTime,labels )
}
*/
