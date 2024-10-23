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
	"sync/atomic"

	"github.com/cloudwego-contrib/cwgo-pkg/config/etcd/utils"

	"github.com/cloudwego-contrib/cwgo-pkg/config/etcd/etcd"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"
)

// WithLimiter sets the limiter config from etcd configuration center.
func WithLimiter(dest string, etcdClient etcd.Client, uniqueID int64, opts utils.Options) server.Option {
	param, err := etcdClient.ServerConfigParam(&etcd.ConfigParamConfig{
		Category:          limiterConfigName,
		ServerServiceName: dest,
	})
	if err != nil {
		panic(err)
	}
	for _, f := range opts.EtcdCustomFunctions {
		f(&param)
	}
	key := param.Prefix + "/" + param.Path
	server.RegisterShutdownHook(func() {
		etcdClient.DeregisterConfig(key, uniqueID)
	})
	return server.WithLimit(initLimitOptions(key, uniqueID, etcdClient))
}

func initLimitOptions(key string, uniqueID int64, etcdClient etcd.Client) *limit.Option {
	var updater atomic.Value
	opt := &limit.Option{}
	opt.UpdateControl = func(u limit.Updater) {
		klog.Debugf("[etcd] %s server etcd limiter updater init, config %v", key, *opt)
		u.UpdateLimit(opt)
		updater.Store(u)
	}
	onChangeCallback := func(restoreDefault bool, data string, parser etcd.ConfigParser) {
		lc := &limiter.LimiterConfig{}

		if !restoreDefault {
			err := parser.Decode(data, lc)
			if err != nil {
				klog.Warnf("[etcd] %s server etcd limiter config: unmarshal data %s failed: %s, skip...", key, data, err)
				return
			}
		}

		opt.MaxConnections = int(lc.ConnectionLimit)
		opt.MaxQPS = int(lc.QPSLimit)
		u := updater.Load()
		if u == nil {
			klog.Warnf("[etcd] %s server etcd limiter config failed as the updater is empty", key)
			return
		}
		if !u.(limit.Updater).UpdateLimit(opt) {
			klog.Warnf("[etcd] %s server etcd limiter config: data %s may do not take affect", key, data)
		}
	}
	etcdClient.RegisterConfigCallback(context.Background(), key, uniqueID, onChangeCallback)
	return opt
}
