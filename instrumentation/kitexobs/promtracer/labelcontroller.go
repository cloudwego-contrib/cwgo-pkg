package prometheus

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	prom "github.com/prometheus/client_golang/prometheus"
)

const (
	labelKeyCaller = semantic.LabelKeyCaller
	labelKeyMethod = semantic.LabelMethod
	labelKeyCallee = semantic.LabelKeyCallee
	labelKeyStatus = semantic.LabelKeyStatus
	labelKeyRetry  = semantic.LabelKeyRetry

	// status
	statusSucceed = "succeed"
	statusError   = "error"

	unknownLabelValue = "unknown"
)

var _ label.LabelControl = PromLabelControl{}

type PromLabelControl struct {
}

func DefaultPromLabelControl() PromLabelControl {
	return PromLabelControl{}
}

func (p PromLabelControl) ProcessAndInjectLabels(ctx context.Context) context.Context {
	return ctx
}

func (p PromLabelControl) ProcessAndExtractLabels(ctx context.Context) []label.CwLabel {
	ri := rpcinfo.GetRPCInfo(ctx)
	extraLabels := make(prom.Labels)
	extraLabels[labelKeyStatus] = statusSucceed
	if ri.Stats().Error() != nil {
		extraLabels[labelKeyStatus] = statusError
	}
	var (
		labels = make(prom.Labels)

		caller = ri.From()
		callee = ri.To()
	)
	labels[labelKeyCaller] = defaultValIfEmpty(caller.ServiceName(), unknownLabelValue)
	labels[labelKeyCallee] = defaultValIfEmpty(callee.ServiceName(), unknownLabelValue)
	labels[labelKeyMethod] = defaultValIfEmpty(callee.Method(), unknownLabelValue)

	labels[labelKeyStatus] = statusSucceed
	if ri.Stats().Error() != nil {
		labels[labelKeyStatus] = statusError
	}

	labels[labelKeyRetry] = "0"
	if retriedCnt, ok := callee.Tag(rpcinfo.RetryTag); ok {
		labels[labelKeyRetry] = retriedCnt
	}
	return label.ToCwLabelFromPromelabel(labels)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
