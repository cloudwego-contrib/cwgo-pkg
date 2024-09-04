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

package label

import (
	prom "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/attribute"
	"strings"
)

type CwLabel struct {
	Key   string
	Value string
}

func ToCwLabelsFromOtels(otelAttributes []attribute.KeyValue) []CwLabel {
	cwLabels := make([]CwLabel, len(otelAttributes))
	for i, attr := range otelAttributes {
		cwLabels[i] = CwLabel{
			Key:   replaceDot(string(attr.Key)),
			Value: attr.Value.AsString(),
		}
	}
	return cwLabels
}

func ToOtelsFromCwLabel(cwLabels []CwLabel) []attribute.KeyValue {
	otelAttributes := make([]attribute.KeyValue, len(cwLabels))
	for i, label := range cwLabels {
		otelAttributes[i] = attribute.String(replaceUnderscore(label.Key), label.Value)
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

func replaceUnderscore(input string) string {
	return strings.ReplaceAll(input, "_", ".")
}
func replaceDot(input string) string {
	return strings.ReplaceAll(input, ".", "_")
}
