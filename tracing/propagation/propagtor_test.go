package propagation

import (
	"context"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego-contrib/obs-opentelemetry/tracing/hertz"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/trace"
	"reflect"
	"testing"
)

func TestExtractHTTP(t *testing.T) {
	ctx := context.Background()
	bags, _ := baggage.Parse("foo=bar")
	ctx = baggage.ContextWithBaggage(ctx, bags)
	ctx = metainfo.WithValue(ctx, "foo", "bar")

	headers := &protocol.RequestHeader{}
	headers.Set("foo", "bar")

	type args struct {
		ctx      context.Context
		c        *hertz.Config
		metadata *protocol.RequestHeader
	}
	tests := []struct {
		name  string
		args  args
		want  baggage.Baggage
		want1 trace.SpanContext
	}{
		{
			name: "extract successful",
			args: args{
				ctx:      ctx,
				c:        hertz.NewConfig([]hertz.Option{}),
				metadata: headers,
			},
			want:  bags,
			want1: trace.SpanContext{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := ExtractHTTP(tt.args.ctx, tt.args.c, tt.args.metadata)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Extract() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Extract() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestInjectHTTP(t *testing.T) {
	cfg := hertz.NewConfig([]hertz.Option{hertz.WithTextMapPropagator(NewPropagator())})

	ctx := context.Background()

	spanContext := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    [16]byte{1},
		SpanID:     [8]byte{2},
		TraceFlags: 0,
		TraceState: trace.TraceState{},
		Remote:     false,
	})

	ctx = trace.ContextWithSpanContext(ctx, spanContext)
	md := &protocol.RequestHeader{}

	type args struct {
		ctx      context.Context
		c        *hertz.Config
		metadata *protocol.RequestHeader
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "inject valid",
			args: args{
				ctx:      ctx,
				c:        cfg,
				metadata: md,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			InjectHTTP(tt.args.ctx, tt.args.c, tt.args.metadata)
			assert.NotEmpty(t, tt.args.metadata)
			assert.Equal(t, "01000000000000000000000000000000-0200000000000000-0", md.Get("b3"))
			assert.Equal(t, "00-01000000000000000000000000000000-0200000000000000-00", md.Get("traceparent"))
		})
	}
}
