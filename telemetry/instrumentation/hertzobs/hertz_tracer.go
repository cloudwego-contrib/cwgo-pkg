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

package hertzobs

import (
	"context"
	"time"

	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/tracer"
	"github.com/cloudwego/hertz/pkg/common/tracer/stats"
)

var _ tracer.Tracer = (*HertzTracer)(nil)

const requestContextKey = "requestContext"

type HertzTracer struct {
	Measure cwmetric.Measure
	cfg     *Config
}

func (h HertzTracer) Start(ctx context.Context, c *app.RequestContext) context.Context {
	ctx = context.WithValue(ctx, requestContextKey, c)
	return h.Measure.ProcessAndInjectLabels(ctx)
}

func (h HertzTracer) Finish(ctx context.Context, c *app.RequestContext) {
	ctx = context.WithValue(ctx, requestContextKey, c)
	ti := c.GetTraceInfo()
	st := ti.Stats()

	if st.Level() == stats.LevelDisabled {
		return
	}

	httpStart := st.GetEvent(stats.HTTPStart)
	if httpStart == nil {
		return
	}
	elapsedTime := float64(st.GetEvent(stats.HTTPFinish).Time().Sub(httpStart.Time())) / float64(time.Millisecond)
	labels := h.Measure.ProcessAndExtractLabels(ctx)

	h.Measure.Inc(ctx, labels)
	h.Measure.Record(ctx, elapsedTime, labels)
}
