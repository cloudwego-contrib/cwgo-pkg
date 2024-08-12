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
