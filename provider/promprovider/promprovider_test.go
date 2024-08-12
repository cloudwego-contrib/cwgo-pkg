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

package promprovider

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"
	"time"
)

const (
	labelKeyCaller = semantic.LabelKeyCaller
	labelKeyMethod = semantic.LabelMethodProm
	labelKeyCallee = semantic.LabelKeyCallee
	labelKeyStatus = semantic.LabelKeyStatus
	labelKeyRetry  = semantic.LabelKeyRetry
)

func TestPromProvider(t *testing.T) {
	registry := prometheus.NewRegistry()
	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "test_counter",
			ConstLabels: prometheus.Labels{"service": "prometheus-test"},
		},
		[]string{"test1", "test2"},
	)
	registry.MustRegister(counter)

	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:        "test_histogram",
			ConstLabels: prometheus.Labels{"service": "prometheus-test"},
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"test1", "test2"},
	)
	registry.MustRegister(histogram)

	mux := http.NewServeMux()
	mux.Handle("/prometheus", promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))

	measure := metric.NewMeasure(metric.NewPromCounter(counter), metric.NewPromRecorder(histogram), nil)
	provider := NewPromProvider(":9090",
		WithRegistry(registry),
		WithMeasure(measure),
		WithServeMux(mux),
	)
	defer provider.Shutdown(context.Background())
	//assert.NoError(t, err, "Failed to register opsProcessed counter")
	labels := []label.CwLabel{
		label.CwLabel{Key: "test1", Value: "abc"},
		label.CwLabel{Key: "test2", Value: "def"},
	}
	// 模拟一些处理
	assert.True(t, measure.Add(context.Background(), 6, labels) == nil)
	assert.True(t, measure.Record(context.Background(), float64(time.Second.Microseconds()), labels) == nil)

	promServerResp, err := http.Get("http://localhost:9090/prometheus")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, promServerResp.StatusCode == http.StatusOK)

	bodyBytes, err := ioutil.ReadAll(promServerResp.Body)
	assert.True(t, err == nil)
	respStr := string(bodyBytes)
	assert.True(t, strings.Contains(respStr, `test_counter{service="prometheus-test",test1="abc",test2="def"} 6`))
	assert.True(t, strings.Contains(respStr, `test_histogram_sum{service="prometheus-test",test1="abc",test2="def"} 1e+06`))

}
