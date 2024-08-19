package prometheus

import (
	"context"
	label2 "github.com/cloudwego-contrib/cwgo-pkg/obs/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/semantic"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	prom "github.com/prometheus/client_golang/prometheus"
)

var _ label2.LabelControl = PromLabelControl{}

type PromLabelControl struct {
}

func DefaultPromLabelControl() PromLabelControl {
	return PromLabelControl{}
}

func (p PromLabelControl) ProcessAndInjectLabels(ctx context.Context) context.Context {
	return ctx
}

func (p PromLabelControl) ProcessAndExtractLabels(ctx context.Context) []label2.CwLabel {
	ri := rpcinfo.GetRPCInfo(ctx)
	extraLabels := make(prom.Labels)
	extraLabels[semantic.LabelKeyStatus] = semantic.StatusSucceed
	if ri.Stats().Error() != nil {
		extraLabels[semantic.LabelKeyStatus] = semantic.StatusError
	}
	var (
		labels = make(prom.Labels)

		caller = ri.From()
		callee = ri.To()
	)
	labels[semantic.LabelRPCCallerKey] = defaultValIfEmpty(caller.ServiceName(), semantic.UnknownLabelValue)
	labels[semantic.LabelRPCCalleeKey] = defaultValIfEmpty(callee.ServiceName(), semantic.UnknownLabelValue)
	labels[semantic.LabelRPCMethodKey] = defaultValIfEmpty(callee.Method(), semantic.UnknownLabelValue)

	labels[semantic.LabelKeyStatus] = semantic.StatusSucceed
	if ri.Stats().Error() != nil {
		labels[semantic.LabelKeyStatus] = semantic.StatusError
	}

	labels[semantic.LabelKeyRetry] = "0"
	if retriedCnt, ok := callee.Tag(rpcinfo.RetryTag); ok {
		labels[semantic.LabelKeyRetry] = retriedCnt
	}
	return label2.ToCwLabelFromPromelabel(labels)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
