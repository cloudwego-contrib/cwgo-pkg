// Copyright 2021 CloudWeGo authors.
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

package adapter

import (
	"context"
	"net/http"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/ut"
	"github.com/cloudwego/hertz/pkg/route"
	"github.com/stretchr/testify/assert"
)

func initServerSentinel(t *testing.T) {
	err := sentinel.InitDefault()
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
	}
	_, err = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "GET:/server_ping",
			Threshold:              1.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               "/api/users/:id",
			Threshold:              0.0,
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			StatIntervalInMs:       1000,
		},
	})
	if err != nil {
		t.Fatalf("Unexpected error: %+v", err)
		return
	}
}

func TestServerSentinelMiddleware(t *testing.T) {
	type args struct {
		opts    []ServerOption
		method  string
		path    string
		reqPath string
		handler func(c context.Context, ctx *app.RequestContext)
	}
	type want struct {
		code int
	}
	tests := []struct {
		name string
		args args
		want want
	}{
		{
			name: "default get",
			args: args{
				opts:    []ServerOption{},
				method:  http.MethodGet,
				path:    "/server_ping",
				reqPath: "/server_ping",
				handler: func(c context.Context, ctx *app.RequestContext) {
					ctx.String(http.StatusOK, "ping")
				},
			},
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name: "customize resource extract",
			args: args{
				opts: []ServerOption{
					WithServerResourceExtractor(func(c context.Context, ctx *app.RequestContext) string {
						return ctx.FullPath()
					}),
				},
				method:  http.MethodGet,
				path:    "/api/users/:id",
				reqPath: "/api/users/123",
				handler: func(c context.Context, ctx *app.RequestContext) {
					ctx.String(http.StatusOK, "ping")
				},
			},
			want: want{
				code: http.StatusTooManyRequests,
			},
		},

		{
			name: "customize block fallback",
			args: args{
				opts: []ServerOption{
					WithServerBlockFallback(func(c context.Context, ctx *app.RequestContext) {
						ctx.String(http.StatusBadRequest, "block")
					}),
				},
				method:  http.MethodGet,
				path:    "/server_ping",
				reqPath: "/server_ping",
				handler: func(c context.Context, ctx *app.RequestContext) {
					ctx.String(http.StatusBadRequest, "")
				},
			},
			want: want{
				code: http.StatusBadRequest,
			},
		},
	}
	initServerSentinel(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := config.NewOptions([]config.Option{})
			router := route.NewEngine(opt)
			router.Use(SentinelServerMiddleware(tt.args.opts...))
			router.Handle(tt.args.method, tt.args.path, tt.args.handler)
			w := ut.PerformRequest(router, tt.args.method, tt.args.reqPath, nil)
			resp := w.Result()
			assert.Equal(t, tt.want.code, resp.StatusCode())
		})
	}
}
