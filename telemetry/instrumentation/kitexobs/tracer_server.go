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

package kitexobs

import (
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"
	"github.com/cloudwego/kitex/server"
)

func NewServerOption(opts ...Option) (server.Option, *Config) {
	tracer := NewServerTracer(opts...)
	kitexTracer, ok := tracer.(*KitexTracer)
	if !ok {
		cfg := NewConfig(opts)
		st := &KitexTracer{}
		serverDurationMeasure, err := cfg.meter.Float64Histogram(semantic.ServerDuration)
		HandleErr(err)
		labelcontrol := NewOtelLabelControl(cfg.tracer, cfg.recordSourceOperation)
		st.Measure = metric.NewMeasure(nil, metric.NewOtelRecorder(serverDurationMeasure), labelcontrol)
		return server.WithTracer(st), cfg
	}
	return server.WithTracer(kitexTracer), kitexTracer.cfg
}
