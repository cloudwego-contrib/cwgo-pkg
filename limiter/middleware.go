package limiter

import (
	"context"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

func NewHertzMiddleware(limiter Limiter) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		doneFunc, err := limiter.Allow()
		if err != nil {
			_ = ctx.AbortWithError(consts.StatusTooManyRequests, err)
			ctx.String(consts.StatusTooManyRequests, ctx.Errors.String())
		} else {
			ctx.Next(c)
			doneFunc()
		}
	}
}
