package metric

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	"go.opentelemetry.io/otel/metric"
)

var _ Counter = &OtelCounter{}

type OtelCounter struct {
	counter metric.Int64Counter
}

func NewOtelCounter(counter metric.Int64Counter) Counter {
	return OtelCounter{
		counter: counter,
	}
}

func (o OtelCounter) Inc(ctx context.Context, labels []label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.counter.Add(ctx, 1, metric.WithAttributes(otelLabel...))
	return nil
}

func (o OtelCounter) Add(ctx context.Context, value int, labels []label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.counter.Add(ctx, int64(value), metric.WithAttributes(otelLabel...))
	return nil
}

var _ Recorder = &OtelRecorder{}

type OtelRecorder struct {
	histogram metric.Float64Histogram
}

func NewOtelRecorder(histogram metric.Float64Histogram) Recorder {
	return OtelRecorder{
		histogram: histogram,
	}
}

func (o OtelRecorder) Record(ctx context.Context, value float64, labels []label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.histogram.Record(ctx, value, metric.WithAttributes(otelLabel...))
	return nil
}
