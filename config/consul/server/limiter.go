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

package server

import (
	"sync/atomic"

	common "github.com/cloudwego-contrib/cwgo-pkg/config/common"

	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/consul"
	"github.com/cloudwego-contrib/cwgo-pkg/config/consul/utils"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"
)

// WithLimiter sets the limiter config from consul configuration center.
func WithLimiter(dest string, consulClient consul.Client, uniqueID int64, opts utils.Options) server.Option {
	param, err := consulClient.ServerConfigParam(&common.ConfigParamConfig{
		Category:          limiterConfigName,
		ServerServiceName: dest,
	})
	if err != nil {
		panic(err)
	}
	for _, f := range opts.ConsulCustomFunctions {
		f(&param)
	}
	key := param.Prefix + "/" + param.Path
	server.RegisterShutdownHook(func() {
		consulClient.DeregisterConfig(key, uniqueID)
	})
	return server.WithLimit(initLimitOptions(param.Type, key, uniqueID, consulClient))
}

func initLimitOptions(kind common.ConfigType, key string, uniqueID int64, consulClient consul.Client) *limit.Option {
	var updater atomic.Value
	opt := &limit.Option{}
	opt.UpdateControl = func(u limit.Updater) {
		klog.Debugf("[consul] %s server consul limiter updater init, config %v", key, *opt)
		u.UpdateLimit(opt)
		updater.Store(u)
	}
	onChangeCallback := func(data string, parser common.ConfigParser) {
		lc := &limiter.LimiterConfig{}

		err := parser.Decode(kind, data, lc)
		if err != nil {
			klog.Warnf("[consul] %s server consul limiter config: unmarshal data %s failed: %s, skip...", key, data, err)
			return
		}

		opt.MaxConnections = int(lc.ConnectionLimit)
		opt.MaxQPS = int(lc.QPSLimit)
		u := updater.Load()
		if u == nil {
			klog.Warnf("[consul] %s server consul limiter config failed as the updater is empty", key)
			return
		}
		if !u.(limit.Updater).UpdateLimit(opt) {
			klog.Warnf("[consul] %s server consul limiter config: data %s may do not take affect", key, data)
		}
	}
	consulClient.RegisterConfigCallback(key, uniqueID, onChangeCallback)
	return opt
}
