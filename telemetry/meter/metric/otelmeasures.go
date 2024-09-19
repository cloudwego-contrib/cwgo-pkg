/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metric

import (
	"context"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	"go.opentelemetry.io/otel/metric"
)

var _ Counter = &OtelCounter{}

type OtelCounter struct {
	counter metric.Int64Counter
}

func NewOtelCounter(counter metric.Int64Counter) Counter {
	if counter == nil {
		return nil
	}
	return OtelCounter{
		counter: counter,
	}
}

func (o OtelCounter) Inc(ctx context.Context, labels ...label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.counter.Add(ctx, 1, metric.WithAttributes(otelLabel...))
	return nil
}

func (o OtelCounter) Add(ctx context.Context, value int, labels ...label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.counter.Add(ctx, int64(value), metric.WithAttributes(otelLabel...))
	return nil
}

var _ Recorder = &OtelRecorder{}

type OtelRecorder struct {
	histogram metric.Float64Histogram
}

func NewOtelRecorder(histogram metric.Float64Histogram) Recorder {
	if histogram == nil {
		return nil
	}
	return &OtelRecorder{
		histogram: histogram,
	}
}

func (o OtelRecorder) Record(ctx context.Context, value float64, labels ...label.CwLabel) error {
	otelLabel := label.ToOtelsFromCwLabel(labels)
	o.histogram.Record(ctx, value, metric.WithAttributes(otelLabel...))
	return nil
}
