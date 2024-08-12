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
	"github.com/cloudwego/kitex/client"
)

func NewClientOption(opts ...Option) (client.Option, *Config) {
	cfg := NewConfig(opts)
	ct := &kitexobs.KitexTracer{}

	clientDurationMeasure, err := cfg.meter.Float64Histogram(semantic.ClientDuration)
	kitexobs.HandleErr(err)
	labelcontrol := kitexobs.NewOtelLabelControl(cfg.tracer, cfg.recordSourceOperation)
	ct.Measure = cwmetric.NewMeasure(nil, cwmetric.NewOtelRecorder(clientDurationMeasure), labelcontrol)
	return client.WithTracer(ct), cfg
}

/*
func (c *clientTracer) Start(ctx context.Context) context.Context {
	ri := rpcinfo.GetRPCInfo(ctx)
	ctx, _ = c.config.tracer.Start(
		ctx,
		spanNaming(ri),
		oteltrace.WithTimestamp(getStartTimeOrNow(ri)),
		oteltrace.WithSpanKind(oteltrace.SpanKindClient),
	)

	return ctx
}

func (c *clientTracer) Finish(ctx context.Context) {
	span := oteltrace.SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		return
	}

	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Stats().Level() == stats.LevelDisabled {
		return
	}

	st := ri.Stats()
	rpcStart := st.GetEvent(stats.RPCStart)
	rpcFinish := st.GetEvent(stats.RPCFinish)
	duration := rpcFinish.Time().Sub(rpcStart.Time())
	elapsedTime := float64(duration) / float64(time.Millisecond)

	attrs := []attribute.KeyValue{
		RPCSystemKitex,
		semconv.RPCMethodKey.String(ri.To().Method()),
		semconv.RPCServiceKey.String(ri.To().ServiceName()),
		RPCSystemKitexRecvSize.Int64(int64(st.RecvSize())),
		RPCSystemKitexSendSize.Int64(int64(st.SendSize())),
		RequestProtocolKey.String(ri.Config().TransportProtocol().String()),
	}

	// The source operation dimension maybe cause high cardinality issues


	if c.config.recordSourceOperation {
		attrs = append(attrs, SourceOperationKey.String(ri.From().Method()))
	}

	span.SetAttributes(attrs...)

	injectStatsEventsToSpan(span, st)

	if panicMsg, panicStack, rpcErr := parseRPCError(ri); rpcErr != nil || len(panicMsg) > 0 {
		recordErrorSpanWithStack(span, rpcErr, panicMsg, panicStack)
	}

	span.End(oteltrace.WithTimestamp(getEndTimeOrNow(ri)))

	metricsAttributes := semantic.ExtractMetricsAttributesFromSpan(span)
	c.Measure.Record(ctx, elapsedTime, label.ToCwLabelsFromOtels(metricsAttributes))
}
*/
