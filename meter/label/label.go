package label

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
)

type CwLabel struct {
	Key   string
	Value string
}

func ToCwLabelsFromOtels(otelAttributes []attribute.KeyValue) []CwLabel {
	cwLabels := make([]CwLabel, len(otelAttributes))
	for i, attr := range otelAttributes {
		cwLabels[i] = CwLabel{
			Key:   string(attr.Key),
			Value: attr.Value.AsString(),
		}
	}
	return cwLabels
}
func ToOtelsFromCwLabel(cwLabels []CwLabel) []attribute.KeyValue {
	otelAttributes := make([]attribute.KeyValue, len(cwLabels))
	for i, label := range cwLabels {
		otelAttributes[i] = attribute.String(label.Key, label.Value)
	}
	return otelAttributes
}

func ToCwLabelFromPromelabel(labels prom.Labels) []CwLabel {
	cwLabels := make([]CwLabel, len(labels))
	i := 0
	for key, value := range labels {
		cwLabels[i] = CwLabel{
			Key:   key,
			Value: value,
		}
		i++
	}
	return cwLabels
}

func ToPromelabelFromCwLabel(labels []CwLabel) prom.Labels {
	promLabels := make(prom.Labels, len(labels))
	for _, label := range labels {
		promLabels[label.Key] = label.Value
	}
	return promLabels
}
