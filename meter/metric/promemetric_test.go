package metric

import (
	"context"
	"github.com/cloudwego-contrib/obs-opentelemetry/log/logging"
	"github.com/cloudwego-contrib/obs-opentelemetry/meter/label"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

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
	prommmetric := NewPrometheusMetrics(counter, histogram)
	assert.Nil(t, prommmetric.Add(ctx, 6, label.ToCwLabelFromPromelabel(labels)))
	assert.Nil(t, prommmetric.Observe(ctx, float64(100*time.Millisecond.Microseconds()), label.ToCwLabelFromPromelabel(labels)))

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
