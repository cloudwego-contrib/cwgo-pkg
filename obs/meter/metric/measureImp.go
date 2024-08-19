package metric

import (
	"context"
	label2 "github.com/cloudwego-contrib/cwgo-pkg/obs/meter/label"
)

var _ Measure = &MeasureImpl{}

type MeasureImpl struct {
	counter  Counter
	recorder Recorder
	label2.LabelControl
}

func (m *MeasureImpl) SetLabelControl(control label2.LabelControl) {
	m.LabelControl = control
}

func NewMeasure(counter Counter, recorder Recorder, labelcontrol label2.LabelControl) Measure {
	return &MeasureImpl{
		counter:      counter,
		recorder:     recorder,
		LabelControl: labelcontrol,
	}
}

// Counter interface implementation
func (m *MeasureImpl) Inc(ctx context.Context, labels []label2.CwLabel) error {
	return m.counter.Inc(ctx, labels)
}

func (m *MeasureImpl) Add(ctx context.Context, value int, labels []label2.CwLabel) error {
	return m.counter.Add(ctx, value, labels)
}

// Recorder interface implementation
func (m *MeasureImpl) Record(ctx context.Context, value float64, labels []label2.CwLabel) error {
	return m.recorder.Record(ctx, value, labels)
}

func (m *MeasureImpl) ProcessAndInjectLabels(ctx context.Context) context.Context {
	return m.LabelControl.ProcessAndInjectLabels(ctx)
}
func (m *MeasureImpl) ProcessAndExtractLabels(ctx context.Context) []label2.CwLabel {
	return m.ProcessAndExtractLabels(ctx)
}
