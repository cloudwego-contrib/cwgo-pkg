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
	"fmt"
	"net"
	"strconv"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoshertz/common"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var _ registry.Registry = (*nacosRegistry)(nil)

type nacosRegistry struct {
	client naming_client.INamingClient
	opts   options.Options
}

func (n *nacosRegistry) Register(info *registry.Info) error {
	if err := n.validRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry-hertz info cwerror: %w", err)
	}

	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return fmt.Errorf("parse registry-hertz info addr cwerror: %w", err)
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry-hertz info port cwerror: %w", err)
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
	}
	// make sure nacos weight >= 0
	weight := 1
	if info.Weight >= 0 {
		weight = info.Weight
	}
	success, err := n.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          host,
		Port:        uint64(p),
		ServiceName: info.ServiceName,
		GroupName:   n.opts.Group,
		ClusterName: n.opts.Cluster,
		Weight:      float64(weight),
		Enable:      true,
		Healthy:     true,
		Ephemeral:   true,
		Metadata:    info.Tags,
	})
	if success {
		hlog.Info("HERTZ: register instance success")
	}
	if err != nil {
		return fmt.Errorf("register instance cwerror: %w", err)
	}

	return nil
}

func (n *nacosRegistry) validRegistryInfo(info *registry.Info) error {
	if info == nil {
		return fmt.Errorf("registry-hertz.Info can not be empty")
	}
	if info.ServiceName == "" {
		return fmt.Errorf("registry-hertz.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return fmt.Errorf("registry-hertz.Info Addr can not be empty")
	}
	return nil
}

func (n *nacosRegistry) Deregister(info *registry.Info) error {
	if err := n.validRegistryInfo(info); err != nil {
		return fmt.Errorf("valid parse registry-hertz info cwerror: %w", err)
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return err
	}
	portInt, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("parse registry-hertz info port cwerror: %w", err)
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
	}
	success, err := n.client.DeregisterInstance(vo.DeregisterInstanceParam{
		Ip:          host,
		Port:        uint64(portInt),
		ServiceName: info.ServiceName,
		GroupName:   n.opts.Group,
		Cluster:     n.opts.Cluster,
		Ephemeral:   true,
	})
	if success {
		hlog.Info("HERTZ: deregister instance success")
	}
	if err != nil {
		return err
	}
	return nil
}

// NewDefaultNacosRegistry create a default service registry-hertz using nacos.
func NewDefaultNacosRegistry(opts ...options.Option) (registry.Registry, error) {
	client, err := common.NewDefaultNacosConfig()
	if err != nil {
		return nil, err
	}
	return NewNacosRegistry(client, opts...), nil
}

// NewNacosRegistry create a new registry-hertz using nacos.
func NewNacosRegistry(client naming_client.INamingClient, opts ...options.Option) registry.Registry {
	opt := options.Options{
		Cluster: "DEFAULT",
		Group:   "DEFAULT_GROUP",
	}
	for _, option := range opts {
		option(&opt)
	}
	return &nacosRegistry{client: client, opts: opt}
}
