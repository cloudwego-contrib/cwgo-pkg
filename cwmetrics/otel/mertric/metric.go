package mertric

import "go.opentelemetry.io/otel/metric"

type Meter struct {
	meter metric.Meter
	cfg   config
}

func NewMeter() {

}
