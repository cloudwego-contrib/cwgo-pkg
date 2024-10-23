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
	"errors"
	"net"
	"time"

	"exampleprom/promWithkitex/kitex_gen/api/echo"

	"exampleprom/promWithkitex/kitex_gen/api"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/instrumentation/otelkitex"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider/promprovider"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	kitexlogrus "github.com/kitex-contrib/obs-opentelemetry/logging/logrus"
)

var _ api.Echo = &EchoImpl{}

type EchoImpl struct{}

// Echo implements the Echo interface.
func (s *EchoImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	klog.CtxDebugf(ctx, "echo called: %s", req.GetMessage())
	nowSec := time.Now().Second()
	if nowSec%3 == 1 {
		klog.CtxErrorf(ctx, "mock error with request message: %s", req.GetMessage())
		return nil, errors.New("mock error")
	}
	if nowSec%3 == 2 {
		klog.CtxErrorf(ctx, "mock panic with request message: %s", req.GetMessage())
		panic("mock panic")
	}
	return &api.Response{Message: req.Message}, nil
}

func main() {
	klog.SetLogger(kitexlogrus.NewLogger())
	// set level as debug when needed, default level is info
	klog.SetLevel(klog.LevelDebug)

	serviceName := "echo"

	p := promprovider.NewPromProvider(
		promprovider.WithServiceName(serviceName),
		promprovider.WithRPCServer(),
	)
	defer p.Shutdown(context.Background())
	p.Serve(":9091", "/kitexserver")
	addr, err := net.ResolveTCPAddr("tcp", ":8181")
	if err != nil {
		panic(err)
	}
	svr := echo.NewServer(
		new(EchoImpl),
		server.WithServiceAddr(addr),
		server.WithSuite(otelkitex.NewServerSuite()),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: serviceName}),
	)
	if err := svr.Run(); err != nil {
		klog.Fatalf("server stopped with error:", err)
	}
}
