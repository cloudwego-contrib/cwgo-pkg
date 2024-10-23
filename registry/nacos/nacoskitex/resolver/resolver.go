// Copyright 2021 CloudWeGo Authors.
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

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/nacos"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type nacosResolver struct {
	cli  naming_client.INamingClient
	opts options.ResolverOptions
}

// NewDefaultNacosResolver create a default service resolver using nacos.
func NewDefaultNacosResolver(opts ...options.ResolverOption) (discovery.Resolver, error) {
	cli, err := nacos.NewDefaultNacosClient()
	if err != nil {
		return nil, err
	}
	return NewNacosResolver(cli, opts...), nil
}

// NewNacosResolver create a service resolver using nacos.
func NewNacosResolver(cli naming_client.INamingClient, opts ...options.ResolverOption) discovery.Resolver {
	op := options.ResolverOptions{
		Cluster: "DEFAULT",
		Group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&op)
	}
	return &nacosResolver{cli: cli, opts: op}
}

// Target return a description for the given target that is suitable for being a key for cache.
func (n *nacosResolver) Target(_ context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve a service info by desc.
func (n *nacosResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	res, err := n.cli.SelectInstances(vo.SelectInstancesParam{
		ServiceName: desc,
		HealthyOnly: true,
		GroupName:   n.opts.Group,
		Clusters:    []string{n.opts.Cluster},
	})
	if err != nil {
		return discovery.Result{}, err
	}
	if len(res) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	instances := make([]discovery.Instance, 0, len(res))
	for _, in := range res {
		if !in.Enable {
			continue
		}
		instances = append(instances, discovery.NewInstance(
			"tcp",
			fmt.Sprintf("%s:%d", in.Ip, in.Port),
			int(in.Weight),
			in.Metadata),
		)
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
func (n *nacosResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name returns the name of the resolver.
func (n *nacosResolver) Name() string {
	return "nacos" + ":" + n.opts.Cluster + ":" + n.opts.Group
}

var _ discovery.Resolver = (*nacosResolver)(nil)
