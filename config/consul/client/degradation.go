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
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/pkg/degradation"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/utils"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
)

func WithDegradation(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := consulClient.ClientConfigParam(&consul.ConfigParamConfig{
		Category:          degradationConfigName,
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
	container := initDegradationOptions(param.Type, key, dest, uniqueID, consulClient)
	return []client.Option{
		client.WithACLRules(container.GetAclRule()),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			consulClient.DeregisterConfig(key, uniqueID)
			return nil
		}),
	}
}

func initDegradationOptions(configType consul.ConfigType, key, dest string, uniqueID int64, consulClient consul.Client) *degradation.DegradationContainer {
	container := degradation.NewDegradationContainer()
	onChangeCallback := func(data string, parser consul.ConfigParser) {
		config := &degradation.DegradationConfig{}
		err := parser.Decode(configType, data, config)
		if err != nil {
			klog.Warnf("[consul] %s server consul degradation config: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}
		container.NotifyPolicyChange(config)
	}
	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)
	return container
}
