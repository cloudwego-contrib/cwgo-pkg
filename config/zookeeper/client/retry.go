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

	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/utils"
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/zookeeper"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/retry"
)

// WithRetryPolicy sets the retry policy from zookeeper configuration center.
func WithRetryPolicy(dest, src string, zookeeperClient zookeeper.Client, opts utils.Options) []client.Option {
	param, err := zookeeperClient.ClientConfigParam(&zookeeper.ConfigParamConfig{
		Category:          retryConfigName,
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
	rc := initRetryContainer(path, uid, dest, zookeeperClient)
	return []client.Option{
		client.WithRetryContainer(rc),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			zookeeperClient.DeregisterConfig(path, uid)
			return rc.Close()
		}),
	}
}

func initRetryContainer(path string, uniqueID int64, dest string, zookeeperClient zookeeper.Client) *retry.Container {
	retryContainer := retry.NewRetryContainerWithPercentageLimit()

	ts := utils.ThreadSafeSet{}

	onChangeCallback := func(restoreDefault bool, data string, parser zookeeper.ConfigParser) {
		// the key is method name, wildcard "*" can match anything.
		rcs := map[string]*retry.Policy{}
		if !restoreDefault && data != "" {
			err := parser.Decode(data, &rcs)
			if err != nil {
				klog.Warnf("[zookeeper] %s client zookeeper retry: unmarshal data %s failed: %s, skip...", path, data, err)
				return
			}
		}

		set := utils.Set{}
		for method, policy := range rcs {
			set[method] = true
			if policy.BackupPolicy != nil && policy.FailurePolicy != nil {
				klog.Warnf("[zookeeper] %s client policy for method %s BackupPolicy and FailurePolicy must not be set at same time",
					dest, method)
				continue
			}
			if policy.BackupPolicy == nil && policy.FailurePolicy == nil {
				klog.Warnf("[zookeeper] %s client policy for method %s BackupPolicy and FailurePolicy must not be empty at same time",
					dest, method)
				continue
			}
			retryContainer.NotifyPolicyChange(method, *policy)
		}

		for _, method := range ts.DiffAndEmplace(set) {
			retryContainer.DeletePolicy(method)
		}
	}

	zookeeperClient.RegisterConfigCallback(context.Background(), path, uniqueID, onChangeCallback)

	return retryContainer
}
