package promtracer

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/obs/semantic"
	"github.com/cloudwego/hertz/pkg/app"
	prom "github.com/prometheus/client_golang/prometheus"
	"strconv"
)

const (
	requestContextKey = "requestContext"
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
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if ok == false {
		return nil
	}
	labels := make(prom.Labels)
	labels[semantic.LabelHttpMethodKey] = defaultValIfEmpty(string(c.Request.Method()), semantic.UnknownLabelValue)
	labels[semantic.LabelKeyStatus] = defaultValIfEmpty(strconv.Itoa(c.Response.Header.StatusCode()), semantic.UnknownLabelValue)
	labels[semantic.LabelPath] = defaultValIfEmpty(c.FullPath(), semantic.UnknownLabelValue)

	return label.ToCwLabelFromPromelabel(labels)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
