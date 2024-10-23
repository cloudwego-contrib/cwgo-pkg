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
	utils "github.com/cloudwego-contrib/cwgo-pkg/config/common"
	"github.com/cloudwego-contrib/cwgo-pkg/config/file/monitor"
	kitexclient "github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/retry"
)

// WithRetryPolicy returns a server.Option that sets the retry policies for the client
func WithRetryPolicy(watcher monitor.ConfigMonitor) []kitexclient.Option {
	rc, keyRetry := initRetryContainer(watcher)
	return []kitexclient.Option{
		kitexclient.WithRetryContainer(rc),
		kitexclient.WithCloseCallbacks(func() error {
			watcher.DeregisterCallback(keyRetry)
			return rc.Close()
		}),
	}
}

// initRetryOptions init the retry container
func initRetryContainer(watcher monitor.ConfigMonitor) (*retry.Container, int64) {
	retryContainer := retry.NewRetryContainerWithPercentageLimit()

	ts := utils.ThreadSafeSet{}

	onChangeCallback := func() {
		// the key is method name, wildcard "*" can match anything.
		config := getFileConfig(watcher)
		if config == nil {
			return // config is nil, do nothing, log will be printed in getFileConfig
		}
		rcs := config.Retry
		set := utils.Set{}

		for method, policy := range rcs {
			set[method] = true

			if policy.BackupPolicy != nil && policy.FailurePolicy != nil {
				klog.Warnf("[local] %s client policy for method %s BackupPolicy and FailurePolicy must not be set at same time",
					watcher.Key(), method)
				continue
			}

			if policy.BackupPolicy == nil && policy.FailurePolicy == nil {
				klog.Warnf("[local] %s client policy for method %s BackupPolicy and FailurePolicy must not be empty at same time",
					watcher.Key(), method)
				continue
			}

			retryContainer.NotifyPolicyChange(method, *policy)
		}

		for _, method := range ts.DiffAndEmplace(set) {
			retryContainer.DeletePolicy(method)
		}
	}

	keyRetry := watcher.RegisterCallback(onChangeCallback)

	return retryContainer, keyRetry
}
