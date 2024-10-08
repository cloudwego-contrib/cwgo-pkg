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
	"github.com/cloudwego/kitex/pkg/retry"
)

// WithRetryPolicy sets the retry policy from consul configuration center.
func WithRetryPolicy(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := consulClient.ClientConfigParam(&common.ConfigParamConfig{
		Category:          retryConfigName,
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
	rc := initRetryContainer(param.Type, key, dest, consulClient, uniqueID)
	return []client.Option{
		client.WithRetryContainer(rc),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			consulClient.DeregisterConfig(key, uniqueID)
			return nil
		}),
		client.WithCloseCallbacks(rc.Close),
	}
}

func initRetryContainer(configType common.ConfigType, key, dest string,
	consulClient consul.Client, uniqueID int64,
) *retry.Container {
	retryContainer := retry.NewRetryContainerWithPercentageLimit()

	ts := common.ThreadSafeSet{}

	onChangeCallback := func(data string, parser common.ConfigParser) {
		// the key is method name, wildcard "*" can match anything.
		rcs := map[string]*retry.Policy{}
		err := parser.Decode(configType, data, &rcs)
		if err != nil {
			klog.Warnf("[consul] %s client consul retry: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}
		set := common.Set{}
		for method, policy := range rcs {
			set[method] = true
			if policy.Enable && policy.BackupPolicy == nil && policy.FailurePolicy == nil {
				klog.Warnf("[consul] %s client policy for method %s BackupPolicy and FailurePolicy must not be empty at same time",
					dest, method)
				continue
			}
			retryContainer.NotifyPolicyChange(method, *policy)
		}

		for _, method := range ts.DiffAndEmplace(set) {
			retryContainer.DeletePolicy(method)
		}
	}

	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)

	return retryContainer
}
