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
	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/apollo"
	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/utils"
	"github.com/cloudwego/kitex/client"
)

type ApolloClientSuite struct {
	apolloClient apollo.Client
	service      string
	client       string
	opts         utils.Options
}

type ClientSuiteOption func(*ApolloClientSuite)

func NewSuite(service, client string, cli apollo.Client,
	options ...utils.Option,
) *ApolloClientSuite {
	client_suite := &ApolloClientSuite{
		service:      service,
		client:       client,
		apolloClient: cli,
	}
	for _, option := range options {
		option.Apply(&client_suite.opts)
	}
	return client_suite
}

func (s *ApolloClientSuite) Options() []client.Option {
	opts := make([]client.Option, 0, 7)
	opts = append(opts, WithRetryPolicy(s.service, s.client, s.apolloClient, s.opts)...)
	opts = append(opts, WithRPCTimeout(s.service, s.client, s.apolloClient, s.opts)...)
	opts = append(opts, WithCircuitBreaker(s.service, s.client, s.apolloClient, s.opts)...)
	return opts
}
