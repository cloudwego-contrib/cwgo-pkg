package meter

import "go.opentelemetry.io/otel/metric"

type Meter struct {
	meter         metric.Meter
	meterProvider metric.MeterProvider
}
