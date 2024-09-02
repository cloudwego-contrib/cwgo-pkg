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

var _ RetryRecorder = &PromRetryRecorder{}

type PromRetryRecorder struct {
	histogram *prometheus.HistogramVec
}

func NewPromRetryRecorder(histogram *prometheus.HistogramVec) *PromRetryRecorder {
	return &PromRetryRecorder{
		histogram: histogram,
	}
}

func (p PromRetryRecorder) RetryRecord(ctx context.Context, value float64, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	histogram, err := p.histogram.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	histogram.Observe(value)
	return nil
}
