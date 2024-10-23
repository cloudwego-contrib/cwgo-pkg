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
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/monitor"
	kitexclient "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpctimeout"
)

// WithRPCTimeout returns a server.Option that sets the timeout provider for the client.
func WithRPCTimeout(watcher monitor.ConfigMonitor) []kitexclient.Option {
	opt, keyRPCTimeout := initRPCTimeout(watcher)
	return []kitexclient.Option{
		kitexclient.WithTimeoutProvider(opt),
		kitexclient.WithCloseCallbacks(func() error {
			watcher.DeregisterCallback(keyRPCTimeout)
			return nil
		}),
	}
}

// initRPCTimeout init the rpc timeout provider
func initRPCTimeout(watcher monitor.ConfigMonitor) (rpcinfo.TimeoutProvider, int64) {
	rpcTimeoutContainer := rpctimeout.NewContainer()

	onChangeCallback := func() {
		// the key is method name, wildcard "*" can match anything.
		config := getFileConfig(watcher)
		if config == nil {
			return // config is nil, do nothing, log will be printed in getFileConfig
		}
		rpcTimeoutContainer.NotifyPolicyChange(config.Timeout)
	}

	keyRPCTimeout := watcher.RegisterCallback(onChangeCallback)
	return rpcTimeoutContainer, keyRPCTimeout
}
