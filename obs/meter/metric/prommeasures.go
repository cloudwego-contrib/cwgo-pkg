package metric

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/meter/label"
	"github.com/prometheus/client_golang/prometheus"
)

var _ Counter = &PromCounter{}

type PromCounter struct {
	counter *prometheus.CounterVec
}

func NewPromCounter(counter *prometheus.CounterVec) *PromCounter {
	return &PromCounter{
		counter: counter,
	}
}
func (p PromCounter) Inc(ctx context.Context, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	counter, err := p.counter.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	counter.Add(float64(1))
	return nil
}

func (p PromCounter) Add(ctx context.Context, value int, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	counter, err := p.counter.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	counter.Add(float64(value))
	return nil
}

var _ Recorder = &PromRecorder{}

type PromRecorder struct {
	histogram *prometheus.HistogramVec
}

func NewPromRecorder(histogram *prometheus.HistogramVec) *PromRecorder {
	return &PromRecorder{
		histogram: histogram,
	}
}

func (p PromRecorder) Record(ctx context.Context, value float64, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	histogram, err := p.histogram.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	histogram.Observe(value)
	return nil
}
