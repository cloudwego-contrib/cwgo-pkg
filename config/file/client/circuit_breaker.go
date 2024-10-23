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
	"strings"

	utils "github.com/cloudwego-contrib/cwgo-pkg/config/common"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/monitor"

	kitexclient "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// WithCircuitBreaker returns a server.Option that sets the circuit breaker for the client
func WithCircuitBreaker(service string, watcher monitor.ConfigMonitor) []kitexclient.Option {
	cbSuite, keyCircuitBreaker := initCircuitBreaker(service, watcher)
	return []kitexclient.Option{
		kitexclient.WithCircuitBreaker(cbSuite),
		kitexclient.WithCloseCallbacks(func() error {
			watcher.DeregisterCallback(keyCircuitBreaker)
			return cbSuite.Close()
		}),
	}
}

// initCircuitBreaker init the circuitbreaker suite
func initCircuitBreaker(service string, watcher monitor.ConfigMonitor) (*circuitbreak.CBSuite, int64) {
	cb := circuitbreak.NewCBSuite(genServiceCBKeyWithRPCInfo)
	lcb := utils.ThreadSafeSet{}

	onChangeCallback := func() {
		set := utils.Set{}
		config := getFileConfig(watcher)
		if config == nil {
			return // config is nil, do nothing, log will be printed in getFileConfig
		}

		for method, config := range config.Circuitbreaker {
			set[method] = true
			key := genServiceCBKey(service, method)
			cb.UpdateServiceCBConfig(key, *config)
		}

		for _, method := range lcb.DiffAndEmplace(set) {
			key := genServiceCBKey(service, method)
			cb.UpdateServiceCBConfig(key, circuitbreak.GetDefaultCBConfig())
		}
	}

	keyCircuitBreaker := watcher.RegisterCallback(onChangeCallback)
	return cb, keyCircuitBreaker
}

func genServiceCBKeyWithRPCInfo(ri rpcinfo.RPCInfo) string {
	if ri == nil {
		return ""
	}
	return genServiceCBKey(ri.To().ServiceName(), ri.To().Method())
}

func genServiceCBKey(toService, method string) string {
	sum := len(toService) + len(method) + 2
	var buf strings.Builder
	buf.Grow(sum)
	buf.WriteString(toService)
	buf.WriteByte('/')
	buf.WriteString(method)
	return buf.String()
}
