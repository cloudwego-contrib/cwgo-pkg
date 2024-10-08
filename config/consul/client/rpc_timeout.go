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
	common "github.com/cloudwego-contrib/cwgo-pkg/config/common"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/utils"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpctimeout"
)

// WithRPCTimeout sets the RPC timeout policy from consul configuration center.
func WithRPCTimeout(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := consulClient.ClientConfigParam(&common.ConfigParamConfig{
		Category:          rpcTimeoutConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	})
	if err != nil {
		panic(err)
	}

	for _, f := range opts.ConsulCustomFunctions {
		f(&param)
	}
	key := param.Prefix + "/" + param.Path
	return []client.Option{
		client.WithTimeoutProvider(initRPCTimeoutContainer(param.Type, key, dest, consulClient, uniqueID)),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			consulClient.DeregisterConfig(key, uniqueID)
			return nil
		}),
	}
}

func initRPCTimeoutContainer(configType common.ConfigType, key, dest string,
	consulClient consul.Client, uniqueID int64,
) rpcinfo.TimeoutProvider {
	rpcTimeoutContainer := rpctimeout.NewContainer()

	onChangeCallback := func(data string, parser common.ConfigParser) {
		configs := map[string]*rpctimeout.RPCTimeout{}
		err := parser.Decode(configType, data, &configs)
		if err != nil {
			klog.Warnf("[consul] %s client consul rpc timeout: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}

		rpcTimeoutContainer.NotifyPolicyChange(configs)
	}

	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)

	return rpcTimeoutContainer
}
