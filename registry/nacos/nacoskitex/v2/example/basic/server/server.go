// Copyright 2024 CloudWeGo Authors.
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

package main

import (
	"context"
	"log"
	"net"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/v2/example/hello/kitex_gen/api"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/v2/example/hello/kitex_gen/api/hello"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/v2/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

type HelloImpl struct{}

func (h *HelloImpl) Echo(_ context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{
		Message: req.Message,
	}
	return
}

func main() {
	r, err := registry.NewDefaultNacosRegistry()
	if err != nil {
		panic(err)
	}
	svr := hello.NewServer(
		new(HelloImpl),
		server.WithRegistry(r),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "Hello"}),
		server.WithServiceAddr(&net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}),
	)
	if err := svr.Run(); err != nil {
		log.Println("server stopped with cwerror:", err)
	} else {
		log.Println("server stopped")
	}
}
