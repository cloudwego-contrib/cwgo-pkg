package metric

import (
	"context"
	"github.com/cloudwego-contrib/obs-opentelemetry/meter/label"
	"github.com/prometheus/client_golang/prometheus"
)

// Labels
const (
	labelKeyCaller = "caller"
	labelKeyCallee = "callee"
	labelKeyMethod = "method"
	labelKeyStatus = "status"
	labelKeyRetry  = "retry"

	// status
	statusSucceed = "succeed"
	statusError   = "error"

	unknownLabelValue = "unknown"
)

var defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}

var _ Metric = PrometheusMetrics{}

type PrometheusMetrics struct {
	handledCounter   *prometheus.CounterVec
	handledHistogram *prometheus.HistogramVec
}

func NewPrometheusMetrics(handledCounter *prometheus.CounterVec, handledHistogram *prometheus.HistogramVec) *PrometheusMetrics {
	return &PrometheusMetrics{
		handledCounter:   handledCounter,
		handledHistogram: handledHistogram,
	}
}

func DelautPrometheusMetrics(countername, servername string) *PrometheusMetrics {
	handledCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: countername,
			Help: "Total number of RPCs completed by the client, regardless of success or failure.",
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	handledHistogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    servername,
			Help:    "Latency (microseconds) of the RPC until it is finished.",
			Buckets: defaultBuckets,
		},
		[]string{labelKeyCaller, labelKeyCallee, labelKeyMethod, labelKeyStatus, labelKeyRetry},
	)
	return &PrometheusMetrics{
		handledCounter:   handledCounter,
		handledHistogram: handledHistogram,
	}
}

func (p PrometheusMetrics) Inc(ctx context.Context, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	counter, err := p.handledCounter.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	counter.Add(float64(1))
	return nil
}

func (p PrometheusMetrics) Add(ctx context.Context, value int, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	counter, err := p.handledCounter.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	counter.Add(float64(value))
	return nil
}

func (p PrometheusMetrics) Observe(ctx context.Context, value float64, labels []label.CwLabel) error {
	pLabel := label.ToPromelabelFromCwLabel(labels)
	histogram, err := p.handledHistogram.GetMetricWith(pLabel)
	if err != nil {
		return err
	}
	histogram.Observe(value)
	return nil
}
