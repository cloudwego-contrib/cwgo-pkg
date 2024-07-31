package propagation

import (
	"context"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel/propagation"
)

const serviceHeader = "x-cw-service-name"

// Metadata is tracing metadata propagator
type Metadata struct {
	propagators propagation.TextMapPropagator
}

var _ propagation.TextMapPropagator = Metadata{}

func NewPropagator() propagation.TextMapPropagator {
	return Metadata{
		propagators: defaultPropagation(),
	}
}

func defaultPropagation() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		b3.New(),
		ot.OT{},
		propagation.Baggage{},
		propagation.TraceContext{},
	)
}

// Inject sets metadata key-values from ctx into the carrier.
func (b Metadata) Inject(ctx context.Context, carrier propagation.TextMapCarrier) {
	b.propagators.Inject(ctx, carrier)
}

// Extract returns a copy of parent with the metadata from the carrier added.
func (b Metadata) Extract(parent context.Context, carrier propagation.TextMapCarrier) context.Context {
	return b.propagators.Extract(parent, carrier)
}

// Fields returns the keys who's values are set with Inject.
func (b Metadata) Fields() []string {
	return []string{serviceHeader}
}
