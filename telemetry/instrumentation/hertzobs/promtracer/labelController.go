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

package promtracer

import (
	"context"
	"strconv"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"github.com/cloudwego/hertz/pkg/app"
	prom "github.com/prometheus/client_golang/prometheus"
)

const (
	requestContextKey = "requestContext"
)

var _ label.LabelControl = PromLabelControl{}

type PromLabelControl struct{}

func DefaultPromLabelControl() PromLabelControl {
	return PromLabelControl{}
}

func (p PromLabelControl) ProcessAndInjectLabels(ctx context.Context) context.Context {
	return ctx
}

func (p PromLabelControl) ProcessAndExtractLabels(ctx context.Context) []label.CwLabel {
	c, ok := ctx.Value(requestContextKey).(*app.RequestContext)
	if !ok {
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
