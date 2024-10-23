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

package consulhertz

import (
	"errors"
	"fmt"
	"net"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/options"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/internal"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/hashicorp/consul/api"
)

const (
	DefaultCheckInterval                       = "5s"
	DefaultCheckTimeout                        = "5s"
	DefaultCheckDeregisterCriticalServiceAfter = "1m"
)

var (
	ErrNilInfo            = errors.New("info is nil")
	ErrMissingServiceName = errors.New("missing service name in consul register")
	ErrMissingAddr        = errors.New("missing addr in consul register")
)

type consulRegistry struct {
	consulClient *api.Client
	opts         options.Options
}

var _ registry.Registry = (*consulRegistry)(nil)

// NewConsulRegister create a new registry-hertz using consul.
func NewConsulRegister(consulClient *api.Client, opts ...options.Option) registry.Registry {
	op := options.Options{
		Check: internal.DefaultCheck(),
	}

	for _, opt := range opts {
		opt(&op)
	}

	return &consulRegistry{consulClient: consulClient, opts: op}
}

// Register register a service to consul.
func (c *consulRegistry) Register(info *registry.Info) error {
	if err := validateRegistryInfo(info); err != nil {
		return fmt.Errorf("validating registry-hertz info failed, err: %w", err)
	}
	host, port, err := internal.ParseAddr(info.Addr)
	if err != nil {
		return fmt.Errorf("parsing addr failed, err: %w", err)
	}

	svcID, err := internal.GetServiceId(info.ServiceName, info.Addr)
	if err != nil {
		return fmt.Errorf("getting service id failed, err: %w", err)
	}

	tags, err := internal.ConvTagMapToSlice(info.Tags)
	if err != nil {
		return err
	}

	svcInfo := &api.AgentServiceRegistration{
		ID:      svcID,
		Name:    info.ServiceName,
		Address: host,
		Port:    port,
		Tags:    tags,
		Weights: &api.AgentWeights{
			Passing: info.Weight,
			Warning: info.Weight,
		},
		Check: c.opts.Check,
	}
	if c.opts.Check != nil {
		c.opts.Check.TCP = net.JoinHostPort(host, fmt.Sprintf("%d", port))
		svcInfo.Check = c.opts.Check
	}

	return c.consulClient.Agent().ServiceRegister(svcInfo)
}

// Deregister deregister a service from consul.
func (c *consulRegistry) Deregister(info *registry.Info) error {
	err := validateRegistryInfo(info)
	if err != nil {
		return fmt.Errorf("validating registry-hertz info failed, err: %w", err)
	}

	svcID, err := getServiceId(info)
	if err != nil {
		return err
	}

	return c.consulClient.Agent().ServiceDeregister(svcID)
}

func validateRegistryInfo(info *registry.Info) error {
	if info == nil {
		return ErrNilInfo
	}
	if info.ServiceName == "" {
		return ErrMissingServiceName
	}
	if info.Addr == nil {
		return ErrMissingAddr
	}

	return nil
}
