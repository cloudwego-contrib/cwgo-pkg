// Copyright 2021 CloudWeGo Authors.
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

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/example/hello/kitex_gen/api"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/example/hello/kitex_gen/api/hello"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/nacos"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type HelloImpl struct{}

func (h *HelloImpl) Echo(_ context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{
		Message: req.Message,
	}
	return
}

func main() {
	// the nacos server config
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}

	// the nacos client config
	cc := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "info",
		CustomLogger:        nacos.NewCustomNacosLogger(), // With a custom Logger，you can use kitex `klog` or other.
	}

	cli, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		panic(err)
	}

	svr := hello.NewServer(
		new(HelloImpl),
		server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{ServiceName: "Hello"}),
		server.WithRegistry(registry.NewNacosRegistry(cli)),
	)
	if err := svr.Run(); err != nil {
		log.Println("server stopped with cwerror:", err)
	} else {
		log.Println("server stopped")
	}
}
