// Copyright 2024 CloudWeGo Authors
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

	consulclient "github.com/cloudwego-contrib/cwgo-pkg/config/consul/client"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/utils"

	"github.com/cloudwego/kitex-examples/kitex_gen/api"
	"github.com/cloudwego/kitex-examples/kitex_gen/api/echo"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
)

type configLog struct{}

func (cl *configLog) Apply(opt *utils.Options) {
	fn := func(k *consul.Key) {
		klog.Infof("consul config %v", k)
	}
	opt.ConsulCustomFunctions = append(opt.ConsulCustomFunctions, fn)
}

func main() {
	consulClient, err := consul.NewClient(consul.Options{})
	if err != nil {
		panic(err)
	}

	cl := &configLog{}

	serviceName := "ServiceName" // your server-side service name
	clientName := "ClientName"   // your client-side service name
	client, err := echo.NewClient(
		serviceName,
		client.WithHostPorts("0.0.0.0:8888"),
		client.WithSuite(consulclient.NewSuite(serviceName, clientName, consulClient, cl)),
	)
	if err != nil {
		log.Fatal(err)
	}
	for {
		req := &api.Request{Message: "my request"}
		resp, err := client.Echo(context.Background(), req)
		if err != nil {
			klog.Errorf("take request cwerror: %v", err)
		} else {
			klog.Infof("receive response %v", resp)
		}
		time.Sleep(time.Second * 10)
	}
}
