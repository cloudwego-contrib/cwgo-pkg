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
	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

var defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}

func TestMetrics(t *testing.T) {
	registry := prom.NewRegistry()
	ctx := context.Background()
	http.Handle("/metrics-demo", promhttp.HandlerFor(registry, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))

	go func() {
		if err := http.ListenAndServe(":9090", nil); err != nil {
			logging.Fatalf("HERTZ: Unable to start a http server, err: %s", err.Error())
		}
	}()

	counter := prom.NewCounterVec(
		prom.CounterOpts{
			Name:        "test_counter",
			ConstLabels: prom.Labels{"service": "prometheus-test"},
		},
		[]string{"test1", "test2"},
	)

	registry.MustRegister(counter)

	histogram := prom.NewHistogramVec(
		prom.HistogramOpts{
			Name:        "test_histogram",
			ConstLabels: prom.Labels{"service": "prometheus-test"},
			Buckets:     defaultBuckets,
		},
		[]string{"test1", "test2"},
	)

	registry.MustRegister(histogram)

	labels := prom.Labels{
		"test1": "abc",
		"test2": "def",
	}
	prommmetric := NewMeasure(NewPromCounter(counter), NewPromRecorder(histogram))
	assert.Nil(t, prommmetric.Add(ctx, 6, label.ToCwLabelFromPromelabel(labels)))
	assert.Nil(t, prommmetric.Record(ctx, float64(100*time.Millisecond.Microseconds()), label.ToCwLabelFromPromelabel(labels)))

	res, err := http.Get("http://localhost:9090/metrics-demo")

	assert.Nil(t, err)

	defer res.Body.Close()

	bodyBytes, err := io.ReadAll(res.Body)

	assert.Nil(t, err)

	bodyStr := string(bodyBytes)

	assert.True(t, strings.Contains(bodyStr, `test_counter{service="prometheus-test",test1="abc",test2="def"} 6`))
	assert.True(t, strings.Contains(bodyStr, `test_histogram_bucket{service="prometheus-test",test1="abc",test2="def",le="50000"} 0`))
	assert.True(t, strings.Contains(bodyStr, `test_histogram_bucket{service="prometheus-test",test1="abc",test2="def",le="100000"} 1`))
}
