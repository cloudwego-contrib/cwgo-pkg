// Copyright 2022 CloudWeGo Authors.
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

package resolver

import (
	"context"
	"fmt"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/servicecomb/options"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/servicecomb/servicecombkitex/servicecomb"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/go-chassis/sc-client"
)

type serviceCombResolver struct {
	cli  *sc.Client
	opts options.ResolverOptions
}

func NewDefaultSCResolver(opts ...options.ResolverOption) (discovery.Resolver, error) {
	client, err := servicecomb.NewDefaultSCClient()
	if err != nil {
		return nil, err
	}
	return NewSCResolver(client, opts...), nil
}

func NewSCResolver(cli *sc.Client, opts ...options.ResolverOption) discovery.Resolver {
	op := options.ResolverOptions{
		AppId:       "DEFAULT",
		VersionRule: "latest",
		ConsumerId:  "",
	}
	for _, option := range opts {
		option(&op)
	}
	return &serviceCombResolver{
		cli:  cli,
		opts: op,
	}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (scr *serviceCombResolver) Target(_ context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve a service info by desc.
func (scr *serviceCombResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	res, err := scr.cli.FindMicroServiceInstances(scr.opts.ConsumerId, scr.opts.AppId, desc, scr.opts.VersionRule, sc.WithoutRevision())
	if err != nil {
		return discovery.Result{}, err
	}
	instances := make([]discovery.Instance, 0, len(res))
	for _, in := range res {
		if in.Status != sc.MSInstanceUP {
			continue
		}
		for _, endPoint := range in.Endpoints {
			instances = append(instances, discovery.NewInstance(
				"tcp",
				endPoint,
				10,
				in.Properties))
		}
	}
	if len(instances) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: instances,
	}, nil
}

// Diff computes the difference between two results.
func (scr *serviceCombResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name returns the name of the resolver.
func (scr *serviceCombResolver) Name() string {
	return "sc-resolver" + ":" + scr.opts.AppId + ":" + scr.opts.VersionRule
}
