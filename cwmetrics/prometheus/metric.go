package prometheus

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"time"
)

var defaultBuckets = []float64{5000, 10000, 25000, 50000, 100000, 250000, 500000, 1000000}

// counterAdd wraps Add of prom.Counter.
func CounterAdd(counterVec *prom.CounterVec, value int, labels prom.Labels) error {
	counter, err := counterVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	counter.Add(float64(value))
	return nil
}

// histogramObserve wraps Observe of prom.Observer.
func HistogramObserve(histogramVec *prom.HistogramVec, value time.Duration, labels prom.Labels) error {
	histogram, err := histogramVec.GetMetricWith(labels)
	if err != nil {
		return err
	}
	histogram.Observe(float64(value.Microseconds()))
	return nil
}
