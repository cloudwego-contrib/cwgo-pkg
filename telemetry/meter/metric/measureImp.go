/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metric

import (
	"context"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
)

var _ Measure = &MeasureImpl{}

type MeasureImpl struct {
	recoders map[string]Recorder
	counters map[string]Counter
}

func NewMeasure(opts ...Option) Measure {
	cfg := newConfig(opts)
	return &MeasureImpl{
		counters: cfg.counter,
		recoders: cfg.recoders,
	}
}

func (m *MeasureImpl) Inc(ctx context.Context, metricType string, labels ...label.CwLabel) error {
	return m.counters[metricType].Inc(ctx, labels...)
}

func (m *MeasureImpl) Add(ctx context.Context, metricType string, value int, labels ...label.CwLabel) error {
	return m.counters[metricType].Add(ctx, value, labels...)
}

// Record Recorder interface implementation
func (m *MeasureImpl) Record(ctx context.Context, metricType string, value float64, labels ...label.CwLabel) error {
	return m.recoders[metricType].Record(ctx, value, labels...)
}
