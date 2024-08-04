// Copyright 2023 CloudWeGo Authors.
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

package metric

import (
	"context"
	"fmt"
	"github.com/cloudwego-contrib/obs-opentelemetry/meter/label"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

var _ Metric = OtelMetrics{}

type OtelMetrics struct {
	counter   metric.Int64Counter
	histogram metric.Float64Histogram
}

func NewOtelMetrics(counter metric.Int64Counter, histogramRecorder metric.Float64Histogram) *OtelMetrics {
	return &OtelMetrics{
		counter:   counter,
		histogram: histogramRecorder,
	}
}

// DeflautOtelMetrics 的默认实现
func DeflautOtelMetrics(meter metric.Meter, countername, histogramname string) *OtelMetrics {
	serverRequestCountMeasure, err := meter.Int64Counter(
		countername,
		metric.WithUnit("count"),
		metric.WithDescription(fmt.Sprint(countername, " count total")),
	)
	handleErr(err)

	serverLatencyMeasure, err := meter.Float64Histogram(
		histogramname,
		metric.WithUnit("ms"),
		metric.WithDescription(fmt.Sprint(histogramname, " duration")),
	)
	handleErr(err)
	ometer := &OtelMetrics{}
	ometer.counter = serverRequestCountMeasure
	ometer.histogram = serverLatencyMeasure
	return ometer
}

func (o OtelMetrics) Inc(ctx context.Context, labels []label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.counter.Add(ctx, 1, metric.WithAttributes(otelLabel...))
	return nil
}

func (o OtelMetrics) Add(ctx context.Context, value int, labels []label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.counter.Add(ctx, int64(value), metric.WithAttributes(otelLabel...))
	return nil
}

func (o OtelMetrics) Record(ctx context.Context, value float64, labels []label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.histogram.Record(ctx, value, metric.WithAttributes(otelLabel...))
	return nil
}
func handleErr(err error) {
	if err != nil {
		otel.Handle(err)
	}
}
