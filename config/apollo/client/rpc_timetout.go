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
	common "github.com/cloudwego-contrib/cwgo-pkg/config/common"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpctimeout"
)

// WithRPCTimeout sets the RPC timeout policy from apollo configuration center.
func WithRPCTimeout(dest, src string, apolloClient apollo.Client,
	opts utils.Options,
) []client.Option {
	param, err := apolloClient.ClientConfigParam(&common.ConfigParamConfig{
		Category:          apollo.RpcTimeoutConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	})
	if err != nil {
		panic(err)
	}
	for _, f := range opts.ApolloCustomFunctions {
		f(&param)
	}

	uniqueID := apollo.GetUniqueID()

	return []client.Option{
		client.WithTimeoutProvider(initRPCTimeoutContainer(param, dest, apolloClient, uniqueID)),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			return apolloClient.DeregisterConfig(param, uniqueID)
		}),
	}
}

func initRPCTimeoutContainer(param apollo.ConfigParam, dest string,
	apolloClient apollo.Client, uniqueID int64,
) rpcinfo.TimeoutProvider {
	rpcTimeoutContainer := rpctimeout.NewContainer()

	onChangeCallback := func(data string, parser common.ConfigParser) {
		configs := map[string]*rpctimeout.RPCTimeout{}
		err := parser.Decode(param.Type, data, &configs)
		if err != nil {
			klog.Warnf("[apollo] %s client apollo rpc timeout: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		rpcTimeoutContainer.NotifyPolicyChange(configs)
	}

	apolloClient.RegisterConfigCallback(param, onChangeCallback, uniqueID)

	return rpcTimeoutContainer
}
