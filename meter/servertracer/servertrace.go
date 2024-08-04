package servertracer

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
)

type TracerHttp interface {
	Start(ctx context.Context, c *app.RequestContext) context.Context
	Finish(ctx context.Context, c *app.RequestContext)
}

type TracerRPC interface {
	Start(ctx context.Context) context.Context
	Finish(ctx context.Context)
}
