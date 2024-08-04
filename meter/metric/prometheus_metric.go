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
	"github.com/cloudwego-contrib/obs-opentelemetry/semantic"
	"github.com/prometheus/client_golang/prometheus"
)

// Labels
const (
	labelKeyCaller = semantic.LabelKeyCaller
	labelKeyMethod = semantic.LabelMethodProm
	labelKeyCallee = semantic.LabelKeyCallee
	labelKeyStatus = semantic.LabelKeyStatus
	labelKeyRetry  = semantic.LabelKeyRetry
)

var defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}

var _ Metric = PrometheusMetrics{}

type PrometheusMetrics struct {
	counter   *prometheus.CounterVec
	histogram *prometheus.HistogramVec
}

func NewPrometheusMetrics(handledCounter *prometheus.CounterVec, handledHistogram *prometheus.HistogramVec) *PrometheusMetrics {
	return &PrometheusMetrics{
		counter:   handledCounter,
		histogram: handledHistogram,
	}
}

func DelautPrometheusMetrics(registry prometheus.Registry, countername, histogramname string) *PrometheusMetrics {
	handledCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: countername,
			Help: fmt.Sprint(countername, " count total"),
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	registry.Register(handledCounter)
	handledHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    histogramname,
			Help:    fmt.Sprint(histogramname, " duration"),
			Buckets: defaultBuckets,
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	registry.Register(handledHistogram)
	return &PrometheusMetrics{
		counter:   handledCounter,
		histogram: handledHistogram,
	}
}

func (p PrometheusMetrics) Inc(ctx context.Context, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	counter, err := p.counter.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	counter.Add(float64(1))
	return nil
}

func (p PrometheusMetrics) Add(ctx context.Context, value int, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	counter, err := p.counter.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	counter.Add(float64(value))
	return nil
}

func (p PrometheusMetrics) Record(ctx context.Context, value float64, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	histogram, err := p.histogram.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	histogram.Observe(value)
	return nil
}
