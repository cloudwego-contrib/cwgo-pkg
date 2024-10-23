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

package client

import (
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/filewatcher"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/monitor"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/parser"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/utils"
	kitexclient "github.com/cloudwego/kitex/client"
)

type FileConfigClientSuite struct {
	watcher monitor.ConfigMonitor
	service string
}

// NewSuite service is the destination service.
func NewSuite(service, key string, watcher filewatcher.FileWatcher, opts ...utils.Option) *FileConfigClientSuite {
	cm, err := monitor.NewConfigMonitor(key, watcher, opts...)
	if err != nil {
		panic(err)
	}

	return &FileConfigClientSuite{
		watcher: cm,
		service: service,
	}
}

// Options return a list client.Option
func (s *FileConfigClientSuite) Options() []kitexclient.Option {
	s.watcher.SetManager(&parser.ClientFileManager{})

	opts := make([]kitexclient.Option, 0, 7)
	opts = append(opts, WithRetryPolicy(s.watcher)...)
	opts = append(opts, WithCircuitBreaker(s.service, s.watcher)...)
	opts = append(opts, WithRPCTimeout(s.watcher)...)
	opts = append(opts, kitexclient.WithCloseCallbacks(func() error {
		s.watcher.Stop()
		return nil
	}))

	if err := s.watcher.Start(); err != nil {
		panic(err)
	}

	return opts
}
