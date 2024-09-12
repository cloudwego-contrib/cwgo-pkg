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

package server

import (
	"context"
	cwutils "github.com/cloudwego-contrib/cwgo-pkg/config/utils"
	"sync/atomic"

	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/utils"
	"github.com/cloudwego-contrib/cwgo-pkg/config/zookeeper/zookeeper"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"
)

// WithLimiter sets the limiter config from zookeeper configuration center.
func WithLimiter(dest string, zookeeperClient zookeeper.Client, opts utils.Options) server.Option {
	param, err := zookeeperClient.ServerConfigParam(&cwutils.ConfigParamConfig{
		Category:          limiterConfigName,
		ServerServiceName: dest,
	})
	if err != nil {
		panic(err)
	}
	for _, f := range opts.ZookeeperCustomFunctions {
		f(&param)
	}
	uid := zookeeper.GetUniqueID()
	path := param.Prefix + "/" + param.Path
	server.RegisterShutdownHook(func() {
		zookeeperClient.DeregisterConfig(path, uid)
	})

	return server.WithLimit(initLimitOptions(path, uid, dest, zookeeperClient))
}

func initLimitOptions(path string, uniqueID int64, dest string, zookeeperClient zookeeper.Client) *limit.Option {
	var updater atomic.Value
	opt := &limit.Option{}
	opt.UpdateControl = func(u limit.Updater) {
		klog.Debugf("[zookeeper] %s server zookeeper limiter updater init, config %v", dest, *opt)
		u.UpdateLimit(opt)
		updater.Store(u)
	}
	onChangeCallback := func(restoreDefault bool, data string, parser cwutils.ConfigParser) {
		lc := &limiter.LimiterConfig{}
		if !restoreDefault && data != "" {
			err := parser.Decode(cwutils.JSON, data, lc)
			if err != nil {
				klog.Warnf("[zookeeper] %s server zookeeper config: unmarshal data %s failed: %s, skip...", dest, data, err)
				return
			}
		}
		opt.MaxConnections = int(lc.ConnectionLimit)
		opt.MaxQPS = int(lc.QPSLimit)
		u := updater.Load()
		if u == nil {
			klog.Warnf("[zookeeper] %s server zookeeper limiter config failed as the updater is empty", dest)
			return
		}
		if !u.(limit.Updater).UpdateLimit(opt) {
			klog.Warnf("[zookeeper] %s server zookeeper limiter config: data %s may do not take affect", dest, data)
		}
	}
	zookeeperClient.RegisterConfigCallback(context.Background(), path, uniqueID, onChangeCallback)
	return opt
}
