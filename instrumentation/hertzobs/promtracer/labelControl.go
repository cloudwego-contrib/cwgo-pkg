package promtracer

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/meter/label"
	"github.com/cloudwego/hertz/pkg/app"
	prom "github.com/prometheus/client_golang/prometheus"
	"strconv"
)

const (
	requestContextKey = "requestContext"
	labelMethod       = "method"
	labelStatusCode   = "statusCode"
	labelPath         = "path"
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
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if ok == false {
		return nil
	}
	labels := make(prom.Labels)
	labels[labelMethod] = defaultValIfEmpty(string(c.Request.Method()), unknownLabelValue)
	labels[labelStatusCode] = defaultValIfEmpty(strconv.Itoa(c.Response.Header.StatusCode()), unknownLabelValue)
	labels[labelPath] = defaultValIfEmpty(c.FullPath(), unknownLabelValue)

	return label.ToCwLabelFromPromelabel(labels)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
