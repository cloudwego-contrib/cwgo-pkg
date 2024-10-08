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
	"strings"

	common "github.com/cloudwego-contrib/cwgo-pkg/config/common"

	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/utils"

	"github.com/cloudwego/kitex/client"
	"github.com/cloudwego/kitex/pkg/circuitbreak"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
)

// WithCircuitBreaker sets the circuit breaker policy from consul configuration center.
func WithCircuitBreaker(dest, src string, consulClient consul.Client, uniqueID int64, opts utils.Options) []client.Option {
	param, err := consulClient.ClientConfigParam(&common.ConfigParamConfig{
		Category:          circuitBreakerConfigName,
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
	cbSuite := initCircuitBreaker(param.Type, key, dest, src, consulClient, uniqueID)

	return []client.Option{
		client.WithCircuitBreaker(cbSuite),
		client.WithCloseCallbacks(func() error {
			// cancel the configuration listener when client is closed.
			consulClient.DeregisterConfig(key, uniqueID)
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
func genServiceCBKeyWithRPCInfo(rpcInfo rpcinfo.RPCInfo) string {
	if rpcInfo == nil {
		return ""
	}
	return genServiceCBKey(rpcInfo.To().ServiceName(), rpcInfo.To().Method())
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

func initCircuitBreaker(configType common.ConfigType, key, dest, src string,
	consulClient consul.Client, uniqueID int64,
) *circuitbreak.CBSuite {
	cb := circuitbreak.NewCBSuite(genServiceCBKeyWithRPCInfo)
	lcb := common.ThreadSafeSet{}

	onChangeCallback := func(data string, parser common.ConfigParser) {
		set := common.Set{}
		configs := map[string]circuitbreak.CBConfig{}
		err := parser.Decode(configType, data, &configs)
		if err != nil {
			klog.Warnf("[consul] %s client consul circuit breaker: unmarshal data %s failed: %s, skip...", key, data, err)
			return
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

	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)

	return cb
}
