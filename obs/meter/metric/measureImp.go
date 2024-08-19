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
	"github.com/cloudwego-contrib/cwgo-pkg/obs/meter/label"
)

var _ Measure = &MeasureImpl{}

type MeasureImpl struct {
	Counter
	Recorder
	label.LabelControl
}

func (m *MeasureImpl) SetLabelControl(control label.LabelControl) {
	m.LabelControl = control
}

func NewMeasure(counter Counter, recorder Recorder, labelcontrol label.LabelControl) Measure {
	return &MeasureImpl{
		Counter:      counter,
		Recorder:     recorder,
		LabelControl: labelcontrol,
	}
}

// Counter interface implementation
func (m *MeasureImpl) Inc(ctx context.Context, labels []label.CwLabel) error {
	return m.Counter.Inc(ctx, labels)
}

func (m *MeasureImpl) Add(ctx context.Context, value int, labels []label.CwLabel) error {
	return m.Counter.Add(ctx, value, labels)
}

// Recorder interface implementation
func (m *MeasureImpl) Record(ctx context.Context, value float64, labels []label.CwLabel) error {
	return m.Recorder.Record(ctx, value, labels)
}

func (m *MeasureImpl) ProcessAndInjectLabels(ctx context.Context) context.Context {
	return m.LabelControl.ProcessAndInjectLabels(ctx)
}
func (m *MeasureImpl) ProcessAndExtractLabels(ctx context.Context) []label.CwLabel {
	return m.LabelControl.ProcessAndExtractLabels(ctx)
}
