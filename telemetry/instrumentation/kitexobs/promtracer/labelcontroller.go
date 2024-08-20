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

package prometheus

import (
	"context"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	prom "github.com/prometheus/client_golang/prometheus"
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
	return label.ToCwLabelFromPromelabel(labels)
}

func defaultValIfEmpty(val, def string) string {
	if val == "" {
		return def
	}
	return val
}
