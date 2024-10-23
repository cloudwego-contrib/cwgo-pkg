/*
 * Copyright 2021 CloudWeGo Authors
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
	"log"
	"time"

	consul "github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulkitex"

	"github.com/cloudwego/kitex/client"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulkitex/example/hello/kitex_gen/api"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/consulkitex/example/hello/kitex_gen/api/hello"
)

func main() {
	r, err := consul.NewConsulResolver("127.0.0.1:8500")
	if err != nil {
		log.Fatal(err)
	}
	c := hello.MustNewClient("hello", client.WithResolver(r), client.WithRPCTimeout(time.Second*3))
	ctx := context.Background()
	for {
		resp, err := c.Echo(ctx, &api.Request{Message: "Hello"})
		if err != nil {
			log.Fatal(err)
		}
		log.Println(resp)
		time.Sleep(time.Second)
	}
}
