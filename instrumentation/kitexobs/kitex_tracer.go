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
	"context"
	cwmetric "github.com/cloudwego-contrib/cwgo-pkg/meter/metric"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/stats"
	"time"
)

var _ stats.Tracer = (*KitexTracer)(nil)

type KitexTracer struct {
	Measure cwmetric.Measure
}

func (s *KitexTracer) Start(ctx context.Context) context.Context {
	return s.Measure.InjectLabels(ctx)
}

func (s *KitexTracer) Finish(ctx context.Context) {
	// rpc info
	ri := rpcinfo.GetRPCInfo(ctx)
	if ri.Stats().Level() == stats.LevelDisabled {
		return
	}

	st := ri.Stats()
	rpcStart := st.GetEvent(stats.RPCStart)
	rpcFinish := st.GetEvent(stats.RPCFinish)
	duration := rpcFinish.Time().Sub(rpcStart.Time())
	elapsedTime := float64(duration) / float64(time.Millisecond)

	labels := s.Measure.ExtractLabels(ctx)
	s.Measure.Inc(ctx, labels)
	s.Measure.Record(ctx, elapsedTime, labels)
}
