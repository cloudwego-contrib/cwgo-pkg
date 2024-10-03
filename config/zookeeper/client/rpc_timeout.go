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
	"context"

	cwutils "github.com/cloudwego-contrib/cwgo-pkg/config/utils"

	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/utils"
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/zookeeper"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/rpctimeout"
)

// WithRPCTimeout sets the RPC timeout policy from zookeeper configuration center.
func WithRPCTimeout(dest, src string, zookeeperClient zookeeper.Client, opts utils.Options) []client.Option {
	param, err := zookeeperClient.ClientConfigParam(&cwutils.ConfigParamConfig{
		Category:          rpcTimeoutConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	})
	if err != nil {
		panic(err)
	}

	for _, f := range opts.ZookeeperCustomFunctions {
		f(&param)
	}

	uid := zookeeper.GetUniqueID()
	path := param.Prefix + "/" + param.Path

	return []client.Option{
		client.WithTimeoutProvider(initRPCTimeoutContainer(path, uid, dest, zookeeperClient)),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			zookeeperClient.DeregisterConfig(path, uid)
			return nil
		}),
	}
}

func initRPCTimeoutContainer(path string, uniqueID int64, dest string, zookeeperClient zookeeper.Client) rpcinfo.TimeoutProvider {
	rpcTimeoutContainer := rpctimeout.NewContainer()

	onChangeCallback := func(restoreDefault bool, data string, parser cwutils.ConfigParser) {
		configs := map[string]*rpctimeout.RPCTimeout{}
		if !restoreDefault {
			err := parser.Decode(cwutils.JSON, data, &configs)
			if err != nil {
				klog.Warnf("[zookeeper] %s client zookeeper rpc timeout: unmarshal data %s failed: %s, skip...", path, data, err)
				return
			}
		}

		rpcTimeoutContainer.NotifyPolicyChange(configs)
	}

	zookeeperClient.RegisterConfigCallback(context.Background(), path, uniqueID, onChangeCallback)

	return rpcTimeoutContainer
}
