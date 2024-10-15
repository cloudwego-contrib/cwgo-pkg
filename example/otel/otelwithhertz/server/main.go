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

package main

import (
	"context"
	"exampleprom/promWithkitex/kitex_gen/api"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/instrumentation/otelhertz"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider/otelprovider"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	hertzlogrus "github.com/hertz-contrib/obs-opentelemetry/logging/logrus"
)

func main() {
	hlog.SetLogger(hertzlogrus.NewLogger())
	hlog.SetLevel(hlog.LevelDebug)

	serviceName := "demo-hertz-server"
	p := otelprovider.NewOpenTelemetryProvider(
		otelprovider.WithServiceName(serviceName),
		// Support setting ExportEndpoint via environment variables: OTEL_EXPORTER_OTLP_ENDPOINT
		otelprovider.WithExportEndpoint("localhost:4317"),
		otelprovider.WithHttpServer(),
		otelprovider.WithInsecure(),
	)
	defer p.Shutdown(context.Background())

	tracer, cfg := otelhertz.NewServerOption()
	h := server.Default(tracer)
	h.Use(otelhertz.ServerMiddleware(cfg))

	h.GET("/ping", func(c context.Context, ctx *app.RequestContext) {
		req := &api.Request{Message: "my request"}

		hlog.CtxDebugf(c, "message received successfully: %s", req.Message)
		ctx.JSON(consts.StatusOK, "resp")
	})

	h.Spin()
}
