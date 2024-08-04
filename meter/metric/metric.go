package metric

import (
	"context"
	"github.com/cloudwego-contrib/obs-opentelemetry/meter/label"
)

type Metric interface {
	Inc(ctx context.Context, labels []label.CwLabel) error
	Add(ctx context.Context, value int, labels []label.CwLabel) error
	Observe(ctx context.Context, value float64, labels []label.CwLabel) error
}
