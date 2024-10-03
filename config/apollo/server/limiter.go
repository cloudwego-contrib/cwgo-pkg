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
	"sync/atomic"

	cwutils "github.com/cloudwego-contrib/cwgo-pkg/config/utils"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/limit"
	"github.com/cloudwego/kitex/pkg/limiter"
	"github.com/cloudwego/kitex/server"

	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/apollo"
	"github.com/cloudwego-contrib/cwgo-pkg/config/apollo/utils"
)

// WithLimiter sets the limiter config from apollo configuration center.
func WithLimiter(dest string, apolloClient apollo.Client,
	opts utils.Options,
) server.Option {
	param, err := apolloClient.ServerConfigParam(&cwutils.ConfigParamConfig{
		Category:          apollo.LimiterConfigName,
		ServerServiceName: dest,
	})
	if err != nil {
		panic(err)
	}
	for _, f := range opts.ApolloCustomFunctions {
		f(&param)
	}
	uniqueID := apollo.GetUniqueID()
	server.RegisterShutdownHook(func() {
		apolloClient.DeregisterConfig(param, uniqueID)
	})
	return server.WithLimit(initLimitOptions(param, dest, apolloClient, uniqueID))
}

func initLimitOptions(param apollo.ConfigParam, dest string, apolloClient apollo.Client, uniqueID int64) *limit.Option {
	var updater atomic.Value
	opt := &limit.Option{}
	opt.UpdateControl = func(u limit.Updater) {
		klog.Debugf("[apollo] %s server apollo limiter updater init, config %v", dest, *opt)
		u.UpdateLimit(opt)
		updater.Store(u)
	}
	onChangeCallback := func(data string, parser cwutils.ConfigParser) {
		lc := &limiter.LimiterConfig{}
		err := parser.Decode(param.Type, data, lc)
		if err != nil {
			klog.Warnf("[apollo] %s server apollo limiter config: unmarshal data %s failed: %s, skip...", dest, data, err)
			return
		}
		opt.MaxConnections = int(lc.ConnectionLimit)
		opt.MaxQPS = int(lc.QPSLimit)
		u := updater.Load()
		if u == nil {
			klog.Warnf("[apollo] %s server apollo limiter config failed as the updater is empty", dest)
			return
		}
		if !u.(limit.Updater).UpdateLimit(opt) {
			klog.Warnf("[apollo] %s server apollo limiter config: data %s may do not take affect", dest, data)
		}
	}

	apolloClient.RegisterConfigCallback(param, onChangeCallback, uniqueID)
	return opt
}
