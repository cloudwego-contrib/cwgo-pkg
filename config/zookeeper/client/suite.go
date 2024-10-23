// Copyright 2023 CloudWeGo Authors
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

package client

import (
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/utils"
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/zookeeper"
	"github.com/cloudwego/kitex/client"
)

const (
	retryConfigName          = "retry"
	rpcTimeoutConfigName     = "rpc_timeout"
	circuitBreakerConfigName = "circuit_break"
	degradationConfigName    = "degradation"
)

// ZookeeperClientSuite zookeeper client config suite, configure retry timeout limit and circuitbreak dynamically from zookeeper.
type ZookeeperClientSuite struct {
	zookeeperClient zookeeper.Client
	service         string
	client          string
	opts            utils.Options
}

// NewSuite service is the destination service name and client is the local identity.
func NewSuite(service, client string, cli zookeeper.Client, opts ...utils.Option) *ZookeeperClientSuite {
	su := &ZookeeperClientSuite{
		service:         service,
		client:          client,
		zookeeperClient: cli,
	}
	for _, f := range opts {
		f.Apply(&su.opts)
	}
	return su
}

// Options return a list client.Option
func (s *ZookeeperClientSuite) Options() []client.Option {
	opts := make([]client.Option, 0, 7)
	opts = append(opts, WithRetryPolicy(s.service, s.client, s.zookeeperClient, s.opts)...)
	opts = append(opts, WithRPCTimeout(s.service, s.client, s.zookeeperClient, s.opts)...)
	opts = append(opts, WithCircuitBreaker(s.service, s.client, s.zookeeperClient, s.opts)...)
	opts = append(opts, WithDegradation(s.service, s.client, s.zookeeperClient, s.opts)...)
	return opts
}
