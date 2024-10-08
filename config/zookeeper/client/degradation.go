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
	"context"

	common "github.com/cloudwego-contrib/cwgo-pkg/config/common"

	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/pkg/degradation"
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/utils"
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/zookeeper"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
)

func WithDegradation(dest, src string, zookeeperClient zookeeper.Client, opts utils.Options) []client.Option {
	param, err := zookeeperClient.ClientConfigParam(&common.ConfigParamConfig{
		Category:          degradationConfigName,
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
	container := initDegradation(path, uid, dest, zookeeperClient)
	return []client.Option{
		client.WithACLRules(container.GetAclRule()),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			zookeeperClient.DeregisterConfig(path, uid)
			return nil
		}),
	}
}

func initDegradation(path string, uniqueID int64, dest string, zookeeperClient zookeeper.Client) *degradation.Container {
	container := degradation.NewContainer()
	onChangeCallback := func(restoreDefault bool, data string, parser common.ConfigParser) {
		config := &degradation.Config{}
		if !restoreDefault {
			err := parser.Decode(common.JSON, data, config)
			if err != nil {
				klog.Warnf("[zookeeper] %s server zookeeper degradation config: unmarshal data %s failed: %s, skip...", path, data, err)
				return
			}
		}
		container.NotifyPolicyChange(config)
	}

	zookeeperClient.RegisterConfigCallback(context.Background(), path, uniqueID, onChangeCallback)

	return container
}
