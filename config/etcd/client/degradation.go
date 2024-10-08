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

	"github.com/cloudwego-contrib/cwgo-pkg/config/etcd/etcd"
	"github.com/cloudwego-contrib/cwgo-pkg/config/etcd/pkg/degradation"
	"github.com/cloudwego-contrib/cwgo-pkg/config/etcd/utils"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
)

func WithDegradation(dest, src string, etcdClient etcd.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := etcdClient.ClientConfigParam(&common.ConfigParamConfig{
		Category:          degradationConfigName,
		ServerServiceName: dest,
		ClientServiceName: src,
	})
	if err != nil {
		panic(err)
	}
	for _, f := range opts.EtcdCustomFunctions {
		f(&param)
	}
	key := param.Prefix + "/" + param.Path
	container := initDegradationOptions(key, dest, uniqueID, etcdClient)
	return []client.Option{
		client.WithACLRules(container.GetAclRule()),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			etcdClient.DeregisterConfig(key, uniqueID)
			return nil
		}),
	}
}

func initDegradationOptions(key, dest string, uniqueID int64, etcdClient etcd.Client) *degradation.Container {
	container := degradation.NewContainer()
	onChangeCallback := func(restoreDefault bool, data string, parser common.ConfigParser) {
		config := &degradation.Config{}
		if !restoreDefault {
			err := parser.Decode(common.JSON, data, config)
			if err != nil {
				klog.Warnf("[etcd] %s server etcd degradation config: unmarshal data %s failed: %s, skip...", key, data, err)
				return
			}
		}
		container.NotifyPolicyChange(config)
	}
	etcdClient.RegisterConfigCallback(context.Background(), key, uniqueID, onChangeCallback)
	return container
}
