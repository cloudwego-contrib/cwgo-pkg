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

package registry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/servicecomb/servicecombkitex/servicecomb"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/go-chassis/cari/discovery"
	"github.com/go-chassis/sc-client"
	"github.com/thoas/go-funk"
)

type scHeartbeat struct {
	cancel      context.CancelFunc
	instanceKey string
}

type options struct {
	appId             string
	versionRule       string
	hostName          string
	heartbeatInterval int32
}

// Option is ServiceComb option.
type Option func(o *options)

// WithAppId with app id option
func WithAppId(appId string) Option {
	return func(o *options) {
		o.appId = appId
	}
}

// WithVersionRule with version rule option
func WithVersionRule(versionRule string) Option {
	return func(o *options) {
		o.versionRule = versionRule
	}
}

// WithHostName with host name option
func WithHostName(hostName string) Option {
	return func(o *options) {
		o.hostName = hostName
	}
}

func WithHeartbeatInterval(second int32) Option {
	return func(o *options) {
		o.heartbeatInterval = second
	}
}

type serviceCombRegistry struct {
	cli         *sc.Client
	opts        options
	lock        *sync.RWMutex
	registryIns map[string]*scHeartbeat
}

// NewDefaultSCRegistry create a new default ServiceComb registry
func NewDefaultSCRegistry(opts ...Option) (registry.Registry, error) {
	client, err := servicecomb.NewDefaultSCClient()
	if err != nil {
		return nil, err
	}
	return NewSCRegistry(client, opts...), nil
}

// NewSCRegistry create a new ServiceComb registry
func NewSCRegistry(client *sc.Client, opts ...Option) registry.Registry {
	op := options{
		appId:             "DEFAULT",
		versionRule:       "1.0.0",
		hostName:          "DEFAULT",
		heartbeatInterval: 5,
	}
	for _, opt := range opts {
		opt(&op)
	}
	return &serviceCombRegistry{
		cli:         client,
		opts:        op,
		lock:        &sync.RWMutex{},
		registryIns: make(map[string]*scHeartbeat),
	}
}

// Register a service info to ServiceComb
func (scr *serviceCombRegistry) Register(info *registry.Info) error {
	if info == nil {
		return errors.New("registry.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry.Info Addr can not be empty")
	}
	host, port, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return fmt.Errorf("parse registry info addr cwerror: %w", err)
	}
	if host == "" || host == "::" {
		host, err = scr.getLocalIpv4Host()
		if err != nil {
			return fmt.Errorf("parse registry info addr cwerror: %w", err)
		}
	}

	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())
	scr.lock.RLock()
	_, ok := scr.registryIns[instanceKey]
	scr.lock.RUnlock()
	if ok {
		return fmt.Errorf("instance{%s} already registered", instanceKey)
	}

	serviceID, err := scr.cli.RegisterService(&discovery.MicroService{
		ServiceName: info.ServiceName,
		AppId:       scr.opts.appId,
		Version:     scr.opts.versionRule,
		Status:      sc.MSInstanceUP,
	})
	if err != nil {
		return fmt.Errorf("register service cwerror: %w", err)
	}

	healthCheck := &discovery.HealthCheck{
		Mode:     "push",
		Interval: 30,
		Times:    3,
	}
	if scr.opts.heartbeatInterval > 0 {
		healthCheck.Interval = scr.opts.heartbeatInterval
	}

	instanceId, err := scr.cli.RegisterMicroServiceInstance(&discovery.MicroServiceInstance{
		ServiceId:   serviceID,
		Endpoints:   []string{host + ":" + port},
		HostName:    scr.opts.hostName,
		HealthCheck: healthCheck,
		Status:      sc.MSInstanceUP,
	})
	if err != nil {
		return fmt.Errorf("register service instance cwerror: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go scr.heartBeat(ctx, serviceID, instanceId)

	scr.lock.Lock()
	defer scr.lock.Unlock()
	scr.registryIns[instanceKey] = &scHeartbeat{
		instanceKey: instanceKey,
		cancel:      cancel,
	}

	return nil
}

// Deregister a service or an instance
func (scr *serviceCombRegistry) Deregister(info *registry.Info) error {
	serviceId, err := scr.cli.GetMicroServiceID(scr.opts.appId, info.ServiceName, scr.opts.versionRule, "")
	if err != nil {
		return fmt.Errorf("get service-id cwerror: %w", err)
	}
	if info.Addr == nil {
		_, err = scr.cli.UnregisterMicroService(serviceId)
		if err != nil {
			return fmt.Errorf("deregister service cwerror: %w", err)
		}
	} else {
		instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())
		scr.lock.RLock()
		insHeartbeat, ok := scr.registryIns[instanceKey]
		scr.lock.RUnlock()
		if !ok {
			return fmt.Errorf("instance{%s} has not registered", instanceKey)
		}

		host, port, err := net.SplitHostPort(info.Addr.String())
		if err != nil {
			return fmt.Errorf("parse deregistry info addr cwerror: %w", err)
		}
		if host == "" || host == "::" {
			host, err = scr.getLocalIpv4Host()
			if err != nil {
				return fmt.Errorf("parse deregistry info addr cwerror: %w", err)
			}
		}

		addr := host + ":" + port

		instanceId := ""
		instances, err := scr.cli.FindMicroServiceInstances("", scr.opts.appId, info.ServiceName, scr.opts.versionRule, sc.WithoutRevision())
		if err != nil {
			return fmt.Errorf("get instances cwerror: %w", err)
		}
		for _, instance := range instances {
			if funk.ContainsString(instance.Endpoints, addr) {
				instanceId = instance.InstanceId
			}
		}
		if instanceId != "" {
			_, err = scr.cli.UnregisterMicroServiceInstance(serviceId, instanceId)
			if err != nil {
				return fmt.Errorf("deregister service cwerror: %w", err)
			}
		}

		scr.lock.Lock()
		insHeartbeat.cancel()
		delete(scr.registryIns, instanceKey)
		scr.lock.Unlock()
	}

	return nil
}

func (scr *serviceCombRegistry) heartBeat(ctx context.Context, serviceId, instanceId string) {
	ticker := time.NewTicker(time.Second * time.Duration(scr.opts.heartbeatInterval))
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			success, err := scr.cli.Heartbeat(serviceId, instanceId)
			if err != nil || !success {
				klog.CtxErrorf(ctx, "beat to ServerComb return cwerror:%+v instance:%v", err, instanceId)
				ticker.Stop()
				return
			}
		}
	}
}

func (scr *serviceCombRegistry) getLocalIpv4Host() (string, error) {
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
