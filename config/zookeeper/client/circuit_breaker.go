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
	"strings"

	common "github.com/cloudwego-contrib/cwgo-pkg/config/common"

	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/utils"
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/zookeeper"
	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// WithCircuitBreaker sets the circuit breaker policy from zookeeper configuration center.
func WithCircuitBreaker(dest, src string, zookeeperClient zookeeper.Client, opts utils.Options) []client.Option {
	param, err := zookeeperClient.ClientConfigParam(&zookeeper.ConfigParamConfig{
		Category:          circuitBreakerConfigName,
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

	cbSuite := initCircuitBreaker(path, uid, dest, zookeeperClient)

	return []client.Option{
		client.WithCircuitBreaker(cbSuite),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			zookeeperClient.DeregisterConfig(path, uid)
			err = cbSuite.Close()
			if err != nil {
				return err
			}
			return nil
		}),
	}
}

// keep consistent when initialising the circuit breaker suit and updating
// the circuit breaker policy.
func genServiceCBKeyWithRPCInfo(ri rpcinfo.RPCInfo) string {
	if ri == nil {
		return ""
	}
	return genServiceCBKey(ri.To().ServiceName(), ri.To().Method())
}

func genServiceCBKey(toService, method string) string {
	sum := len(toService) + len(method) + 2
	var buf strings.Builder
	buf.Grow(sum)
	buf.WriteString(toService)
	buf.WriteByte('/')
	buf.WriteString(method)
	return buf.String()
}

func initCircuitBreaker(path string, uniqueID int64, dest string, zookeeperClient zookeeper.Client) *circuitbreak.CBSuite {
	cb := circuitbreak.NewCBSuite(genServiceCBKeyWithRPCInfo)
	lcb := common.ThreadSafeSet{}

	onChangeCallback := func(restoreDefault bool, data string, parser zookeeper.ConfigParser) {
		set := common.Set{}
		configs := map[string]circuitbreak.CBConfig{}

		if !restoreDefault {
			err := parser.Decode(data, &configs)
			if err != nil {
				klog.Warnf("[zookeeper] %s client zookeeper circuit breaker: unmarshal data %s failed: %s, skip...", path, data, err)
				return
			}
		}

		for method, config := range configs {
			set[method] = true
			key := genServiceCBKey(dest, method)
			cb.UpdateServiceCBConfig(key, config)
		}

		for _, method := range lcb.DiffAndEmplace(set) {
			key := genServiceCBKey(dest, method)
			// For deleted method configs, set to default policy
			cb.UpdateServiceCBConfig(key, circuitbreak.GetDefaultCBConfig())
		}
	}

	zookeeperClient.RegisterConfigCallback(context.Background(), path, uniqueID, onChangeCallback)

	return cb
}
