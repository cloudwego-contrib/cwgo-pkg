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

package servicecombhertz

import (
	"context"
	"errors"
	"fmt"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/servicecomb/options"
	"net"
	"sync"
	"time"

	"github.com/go-chassis/sc-client"
	"github.com/thoas/go-funk"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/go-chassis/cari/discovery"
)

var _ registry.Registry = (*serviceCombRegistry)(nil)

type scHeartbeat struct {
	cancel      context.CancelFunc
	instanceKey string
}

type serviceCombRegistry struct {
	cli         *sc.Client
	opts        options.Options
	lock        *sync.RWMutex
	registryIns map[string]*scHeartbeat
}

// NewDefaultSCRegistry create a new default ServiceComb registry-hertz
func NewDefaultSCRegistry(endPoints []string, opts ...options.Option) (registry.Registry, error) {
	client, err := sc.NewClient(sc.Options{
		Endpoints: endPoints,
	})
	if err != nil {
		return nil, err
	}
	return NewSCRegistry(client, opts...), nil
}

// NewSCRegistry create a new ServiceComb registry-hertz
func NewSCRegistry(client *sc.Client, opts ...options.Option) registry.Registry {
	op := options.Options{
		AppId:             "DEFAULT",
		VersionRule:       "1.0.0",
		HostName:          "DEFAULT",
		HeartbeatInterval: 5,
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
	err := scr.validRegistryInfo(info)
	if err != nil {
		return err
	}
	addr, err := scr.parseAddr(info.Addr.String())
	if err != nil {
		return err
	}
	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, addr)
	scr.lock.RLock()
	_, ok := scr.registryIns[instanceKey]
	scr.lock.RUnlock()
	if ok {
		return fmt.Errorf("instance{%s} already registered", instanceKey)
	}

	serviceID, err := scr.cli.RegisterService(&discovery.MicroService{
		ServiceName: info.ServiceName,
		AppId:       scr.opts.AppId,
		Version:     scr.opts.VersionRule,
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
	if scr.opts.HeartbeatInterval > 0 {
		healthCheck.Interval = scr.opts.HeartbeatInterval
	}

	instanceId, err := scr.cli.RegisterMicroServiceInstance(&discovery.MicroServiceInstance{
		ServiceId:   serviceID,
		Endpoints:   []string{addr},
		HostName:    scr.opts.HostName,
		HealthCheck: healthCheck,
		Status:      sc.MSInstanceUP,
		Properties:  info.Tags,
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
	err := scr.validRegistryInfo(info)
	if err != nil {
		return err
	}

	addr, err := scr.parseAddr(info.Addr.String())
	if err != nil {
		return err
	}

	serviceId, err := scr.cli.GetMicroServiceID(scr.opts.AppId, info.ServiceName, scr.opts.VersionRule, "")
	if err != nil {
		return fmt.Errorf("get service-id cwerror: %w", err)
	}

	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, addr)
	scr.lock.RLock()
	insHeartbeat, ok := scr.registryIns[instanceKey]
	scr.lock.RUnlock()
	if !ok {
		return fmt.Errorf("instance{%s} has not registered", instanceKey)
	}

	instanceId := ""
	instances, err := scr.cli.FindMicroServiceInstances("", scr.opts.AppId, info.ServiceName, scr.opts.VersionRule, sc.WithoutRevision())
	if err != nil {
		return fmt.Errorf("get instances cwerror: %w", err)
	}
	for _, instance := range instances {
		if funk.ContainsString(instance.Endpoints, addr) {
			instanceId = instance.InstanceId
		}
	}
	if instanceId != "" {
		// unregister is too slow to take effect, update status to down first.
		_, err = scr.cli.UpdateMicroServiceInstanceStatus(serviceId, instanceId, sc.MSIinstanceDown)
		if err != nil {
			return fmt.Errorf("down service cwerror: %w", err)
		}
		_, err = scr.cli.UnregisterMicroServiceInstance(serviceId, instanceId)
		if err != nil {
			return fmt.Errorf("deregister service cwerror: %w", err)
		}
	}

	scr.lock.Lock()
	insHeartbeat.cancel()
	delete(scr.registryIns, instanceKey)
	scr.lock.Unlock()
	return nil
}

func (scr *serviceCombRegistry) heartBeat(ctx context.Context, serviceId, instanceId string) {
	ticker := time.NewTicker(time.Second * time.Duration(scr.opts.HeartbeatInterval))
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			success, err := scr.cli.Heartbeat(serviceId, instanceId)
			if err != nil || !success {
				hlog.CtxErrorf(ctx, "HERTZ: beat to ServerComb return cwerror:%+v instance:%v", err, instanceId)
				ticker.Stop()
				return
			}
		}
	}
}

func (scr *serviceCombRegistry) validRegistryInfo(info *registry.Info) error {
	if info == nil {
		return errors.New("registry-hertz.Info can not be empty")
	}
	if info.ServiceName == "" {
		return errors.New("registry-hertz.Info ServiceName can not be empty")
	}
	if info.Addr == nil {
		return errors.New("registry-hertz.Info Addr can not be empty")
	}
	return nil
}

func (scr *serviceCombRegistry) parseAddr(s string) (string, error) {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return "", fmt.Errorf("parse addr cwerror: %w", err)
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
		if host == utils.UNKNOWN_IP_ADDR {
			return "", errors.New("get local ip cwerror")
		}
	}

	return net.JoinHostPort(host, port), nil
}
