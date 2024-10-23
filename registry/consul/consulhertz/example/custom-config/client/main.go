// Copyright 2022 CloudWeGo Authors.
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
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulhertz"

	"github.com/cloudwego/hertz/pkg/app/client"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/cloudwego/hertz/pkg/app/client/loadbalance"
	"github.com/cloudwego/hertz/pkg/app/middlewares/client/sd"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/hashicorp/consul/api"
)

var localIP = "your ip"

type Example struct {
	A int `json:"a"`
	B int `json:"b"`
}

func main() {
	// build a consul client
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "127.0.0.1:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	// build a consul resolver with the consul client
	r := consulhertz.NewConsulResolver(consulClient)

	discoveryWithSD(r)
	discoveryWithTag(r)
	discoveryWithCustomizedAddr(r)
	discoveryWithLoadBalanceOptions(r)
	discoveryThenUsePostMethod(r)
}

func discoveryWithSD(r discovery.Resolver) {
	fmt.Println("simply discovery:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://custom-config-demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

func discoveryWithTag(r discovery.Resolver) {
	fmt.Println("discovery with tag:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://custom-config-demo/ping", config.WithSD(true), config.WithTag("key1", "val1"))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

func discoveryWithCustomizedAddr(r discovery.Resolver) {
	fmt.Println("discovery with customizedAddr:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}

	cli.Use(sd.Discovery(r, sd.WithCustomizedAddrs(net.JoinHostPort(localIP, "5001"))))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://custom-config-demo/ping", config.WithSD(true), config.WithTag("key1", "val1"))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

func discoveryWithLoadBalanceOptions(r discovery.Resolver) {
	fmt.Println("discovery with loadBalanceOptions:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r, sd.WithLoadBalanceOptions(loadbalance.NewWeightedBalancer(), loadbalance.Options{
		RefreshInterval: 5 * time.Second,
		ExpireInterval:  15 * time.Second,
	})))
	for i := 0; i < 10; i++ {
		status, body, err := cli.Get(context.Background(), nil, "http://custom-config-demo/ping", config.WithSD(true))
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", status, string(body))
	}
}

func discoveryThenUsePostMethod(r discovery.Resolver) {
	fmt.Println("discovery and use post method to send request:")
	cli, err := client.NewClient()
	if err != nil {
		panic(err)
	}
	cli.Use(sd.Discovery(r))

	for i := 0; i < 10; i++ {
		// set request config、method、request uri.
		req := protocol.AcquireRequest()
		req.SetOptions(config.WithSD(true))
		req.SetMethod("POST")
		req.SetRequestURI("http://custom-config-demo/ping")
		t := Example{A: i, B: i}
		bytes, _ := json.Marshal(t)
		// set body and content type
		req.SetBody(bytes)
		req.Header.SetContentTypeBytes([]byte("application/json"))
		resp := protocol.AcquireResponse()
		// send request
		err := cli.Do(context.Background(), req, resp)
		if err != nil {
			hlog.Fatal(err)
		}
		hlog.Infof("code=%d,body=%s", resp.StatusCode(), string(resp.Body()))
	}
}
