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

package registry

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/nacos"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type nacosRegistry struct {
	cli  naming_client.INamingClient
	opts options.Options
}

// NewDefaultNacosRegistry create a default service registry using nacos.
func NewDefaultNacosRegistry(opts ...options.Option) (registry.Registry, error) {
	cli, err := nacos.NewDefaultNacosClient()
	if err != nil {
		return nil, err
	}
	return NewNacosRegistry(cli, opts...), nil
}

// NewNacosRegistry create a new registry using nacos.
func NewNacosRegistry(cli naming_client.INamingClient, opts ...options.Option) registry.Registry {
	op := options.Options{
		Cluster: "DEFAULT",
		Group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&op)
	}
	return &nacosRegistry{cli: cli, opts: op}
}

var _ registry.Registry = (*nacosRegistry)(nil)

// Register service info to nacos.
func (n *nacosRegistry) Register(info *registry.Info) error {
	if err := n.validateRegistryInfo(info); err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return fmt.Errorf("parse registry info addr cwerror: %w", err)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry info port cwerror: %w", err)
	}
	if host == "" || host == "::" {
		host, err = n.getLocalIpv4Host()
		if err != nil {
			return fmt.Errorf("parse registry info addr cwerror: %w", err)
		}
	}
	// make sure nacos weight >= 0
	weight := 1
	if info.Weight >= 0 {
		weight = info.Weight
	}
	_, e := n.cli.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		Weight:      float64(weight),
		Enable:      true,
		Healthy:     true,
		Metadata:    mergeTags(info.Tags, nacos.Tags),
		GroupName:   n.opts.Group,
		ClusterName: n.opts.Cluster,
		Ephemeral:   true,
	})
	if e != nil {
		return fmt.Errorf("register instance cwerror: %w", e)
	}
	return nil
}

func (n *nacosRegistry) getLocalIpv4Host() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addr {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("not found ipv4 address")
}

func (n *nacosRegistry) validateRegistryInfo(info *registry.Info) error {
	if info == nil {
		return errors.New("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry.Info Addr can not be empty")
	}

	return nil
}

// Deregister a service info from nacos.
func (n *nacosRegistry) Deregister(info *registry.Info) error {
	if err := n.validateRegistryInfo(info); err != nil {
		return err
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return err
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry info port cwerror: %w", err)
	}
	if host == "" || host == "::" {
		host, err = n.getLocalIpv4Host()
		if err != nil {
			return fmt.Errorf("parse registry info addr cwerror: %w", err)
		}
	}
	if _, err = n.cli.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		Ephemeral:   true,
		GroupName:   n.opts.Group,
		Cluster:     n.opts.Cluster,
	}); err != nil {
		return err
	}
	return nil
}

// should not modify the source data.
func mergeTags(ts ...map[string]string) map[string]string {
	if len(ts) == 0 {
		return nil
	}
	if len(ts) == 1 {
		return ts[0]
	}
	tags := map[string]string{}
	for _, t := range ts {
		for k, v := range t {
			tags[k] = v
		}
	}
	return tags
}
