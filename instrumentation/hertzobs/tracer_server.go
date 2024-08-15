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

package hertzobs

import (
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/semantic"
	"go.opentelemetry.io/otel/metric"

	"github.com/cloudwego/hertz/pkg/app/server"
	serverconfig "github.com/cloudwego/hertz/pkg/common/config"
)

func NewServerTracer(opts ...Option) (serverconfig.Option, *Config) {
	cfg := NewConfig(opts)
	st := &HertzTracer{}

	serverRequestCountMeasure, err := cfg.meter.Int64Counter(
		semantic.ServerRequestCount,
		metric.WithUnit("count"),
		metric.WithDescription("measures Incoming request count total"),
	)
	handleErr(err)

	serverLatencyMeasure, err := cfg.meter.Float64Histogram(
		semantic.ServerLatency,
		metric.WithUnit("ms"),
		metric.WithDescription("measures th incoming end to end duration"),
	)
	handleErr(err)
	labelControl := NewOtelLabelControl(cfg.tracer, cfg.shouldIgnore, cfg.serverHttpRouteFormatter)
	st.Measure = cwmetric.NewMeasure(cwmetric.NewOtelCounter(serverRequestCountMeasure), cwmetric.NewOtelRecorder(serverLatencyMeasure), labelControl)

	return server.WithTracer(st), cfg
}
