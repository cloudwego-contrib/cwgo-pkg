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

package server

import (
	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/apollo"
	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/utils"
	"github.com/cloudwego/kitex/server"
)

// ApolloServerSuite apollo server config suite, configure limiter config dynamically from apollo.
type ApolloServerSuite struct {
	apolloClient apollo.Client
	service      string
	opts         utils.Options
}

// NewSuite service is the destination service.
func NewSuite(service string, cli apollo.Client, options ...utils.Option,
) *ApolloServerSuite {
	server_suite := &ApolloServerSuite{
		service:      service,
		apolloClient: cli,
	}
	for _, option := range options {
		option.Apply(&server_suite.opts)
	}
	return server_suite
}

// Options return a list client.Option
func (s *ApolloServerSuite) Options() []server.Option {
	opts := make([]server.Option, 0, 2)
	opts = append(opts, WithLimiter(s.service, s.apolloClient, s.opts))
	return opts
}
