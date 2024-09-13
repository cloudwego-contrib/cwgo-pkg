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

package nacos

import (
	"context"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"
	"net"
	"net/url"
	"strconv"
	"strings"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/common"
	"github.com/cloudwego/hertz/pkg/app/client/discovery"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var _ discovery.Resolver = (*nacosResolver)(nil)

type nacosResolver struct {
	client naming_client.INamingClient
	opts   options.ResolverOptions
}

func (n *nacosResolver) Target(_ context.Context, target *discovery.TargetInfo) string {
	var metadata strings.Builder

	// Set serviceName and metadata to desc
	tags := target.Tags
	if len(tags) == 0 {
		return target.Host
	}

	metadata.WriteString(target.Host)
	metadata.WriteString("?")
	values := url.Values{}
	for k, v := range tags {
		values.Add(k, v)
	}
	metadata.WriteString(values.Encode())
	return metadata.String()
}

func (n *nacosResolver) Resolve(_ context.Context, desc string) (discovery.Result, error) {
	var metadata map[string]string
	serviceName := desc

	// Get serviceName and metadata from desc
	if strings.Contains(desc, "?") {
		queries, _ := url.Parse(desc)
		tags, _ := url.ParseQuery(queries.Query().Encode())

		result := make(map[string]string)
		for key, value := range tags {
			result[key] = value[0]
		}
		metadata = result
		serviceName = strings.Split(desc, "?")[0]
	}

	res, err := n.client.SelectInstances(vo.SelectInstancesParam{
		ServiceName: serviceName,
		HealthyOnly: true,
		GroupName:   n.opts.Group,
		Clusters:    []string{n.opts.Cluster},
	})
	if err != nil {
		return discovery.Result{}, err
	}
	instances := make([]discovery.Instance, 0, len(res))
	for _, ins := range res {
		if !ins.Enable || (len(metadata) > 0 && !compareMaps(ins.Metadata, metadata)) {
			continue
		}

		formatPort := strconv.FormatUint(ins.Port, 10)
		instances = append(instances,
			discovery.NewInstance(
				"tcp",
				net.JoinHostPort(ins.Ip, formatPort),
				int(ins.Weight), ins.Metadata,
			),
		)
	}

	return discovery.Result{
		CacheKey:  desc,
		Instances: instances,
	}, nil
}

func (n *nacosResolver) Name() string {
	return "nacos" + ":" + n.opts.Cluster + ":" + n.opts.Group
}

// NewDefaultNacosResolver create a default service resolver using nacos.
func NewDefaultNacosResolver(opts ...options.ResolverOption) (discovery.Resolver, error) {
	client, err := common.NewDefaultNacosConfig()
	if err != nil {
		return nil, err
	}
	return NewNacosResolver(client, opts...), nil
}

// NewNacosResolver create a service resolver using nacos.
func NewNacosResolver(cli naming_client.INamingClient, opts ...options.ResolverOption) discovery.Resolver {
	opt := options.ResolverOptions{
		Cluster: "DEFAULT",
		Group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&opt)
	}
	return &nacosResolver{client: cli, opts: opt}
}

// compareMaps compares two maps regardless of nil or empty
func compareMaps(m1, m2 map[string]string) bool {
	// if both maps are nil, they are equal
	if m1 == nil && m2 == nil {
		return true
	}
	// if the lengths are different, the maps are not equal
	if len(m1) != len(m2) {
		return false
	}
	// iterate over the keys of m1 and check if they exist in m2 with the same value
	for k, v := range m1 {
		if v2, ok := m2[k]; !ok || v != v2 {
			return false
		}
	}
	// return true if no differences are found
	return true
}
