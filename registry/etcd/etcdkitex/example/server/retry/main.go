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
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/etcd/etcdkitex"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/etcd/etcdkitex/retry"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api/hello"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
)

type HelloImpl struct{}

func (h *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	resp = &api.Response{
		Message: req.Message,
	}
	return
}

func main() {
	retryConfig := retry.NewRetryConfig(
		retry.WithMaxAttemptTimes(10),
		retry.WithObserveDelay(20*time.Second),
		retry.WithRetryDelay(5*time.Second),
	)
	r, err := etcdkitex.NewEtcdRegistryWithRetry([]string{"127.0.0.1:2379"}, retryConfig)
	if err != nil {
		log.Fatal(err)
	}
	server := hello.NewServer(new(HelloImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "Hello",
	}))
	err = server.Run()
	if err != nil {
		log.Fatal(err)
	}
}
