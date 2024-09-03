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

package eurekahertz

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/cwerror"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/internal"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/hudl/fargo"
)

var _ registry.Registry = (*eurekaRegistry)(nil)

type eurekaHeartbeat struct {
	cancel      context.CancelFunc
	instanceKey string
}

type eurekaRegistry struct {
	eurekaConn       *fargo.EurekaConnection
	lock             *sync.RWMutex
	registryIns      map[string]*eurekaHeartbeat
	heatBeatInterval time.Duration
}

// NewEurekaRegistry creates a eureka registry-etcdhertz.
func NewEurekaRegistry(servers []string, heatBeatInterval time.Duration) *eurekaRegistry {
	conn := fargo.NewConn(servers...)

	return &eurekaRegistry{
		eurekaConn:       &conn,
		registryIns:      make(map[string]*eurekaHeartbeat),
		lock:             &sync.RWMutex{},
		heatBeatInterval: heatBeatInterval,
	}
}

// NewEurekaRegistryFromConfig creates a eureka registry-etcdhertz.
func NewEurekaRegistryFromConfig(config fargo.Config, heatBeatInterval time.Duration) *eurekaRegistry {
	conn := fargo.NewConnFromConfig(config)

	return &eurekaRegistry{
		eurekaConn:       &conn,
		registryIns:      make(map[string]*eurekaHeartbeat),
		lock:             &sync.RWMutex{},
		heatBeatInterval: heatBeatInterval,
	}
}

// NewEurekaRegistryFromConn creates a eureka registry-etcdhertz.
func NewEurekaRegistryFromConn(conn fargo.EurekaConnection, heatBeatInterval time.Duration) *eurekaRegistry {
	return &eurekaRegistry{
		eurekaConn:       &conn,
		registryIns:      make(map[string]*eurekaHeartbeat),
		lock:             &sync.RWMutex{},
		heatBeatInterval: heatBeatInterval,
	}
}

// Deregister deregister a server with given registry-etcdhertz info.
func (e *eurekaRegistry) Deregister(info *registry.Info) error {
	instance, err := e.eurekaInstance(info)
	if err != nil {
		return err
	}

	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())

	e.lock.RLock()
	insHeartbeat, ok := e.registryIns[instanceKey]
	e.lock.RUnlock()
	if !ok {
		return fmt.Errorf("instance{%s} has not registered", instanceKey)
	}

	if err = e.eurekaConn.DeregisterInstance(instance); err != nil {
		return err
	}

	e.lock.Lock()
	insHeartbeat.cancel()
	delete(e.registryIns, instanceKey)
	e.lock.Unlock()

	return nil
}

// Register a server with given registry-etcdhertz info.
func (e *eurekaRegistry) Register(info *registry.Info) error {
	instance, err := e.eurekaInstance(info)
	if err != nil {
		return err
	}

	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())

	e.lock.RLock()
	_, ok := e.registryIns[instanceKey]
	e.lock.RUnlock()
	if ok {
		return fmt.Errorf("instance{%s} already registered", instanceKey)
	}

	if err = e.eurekaConn.RegisterInstance(instance); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	go e.heartBeat(ctx, instance)

	e.lock.Lock()
	defer e.lock.Unlock()
	e.registryIns[instanceKey] = &eurekaHeartbeat{
		instanceKey: instanceKey,
		cancel:      cancel,
	}

	return nil
}

func (e *eurekaRegistry) eurekaInstance(info *registry.Info) (*fargo.Instance, error) {
	if info == nil {
		return nil, cwerror.ErrNilInfo
	}
	if info.Addr == nil {
		return nil, cwerror.ErrNilAddr
	}
	if len(info.ServiceName) == 0 {
		return nil, cwerror.ErrEmptyServiceName
	}

	host, portStr, err := net.SplitHostPort(info.Addr.String())
	if err != nil {
		return nil, err
	}
	if portStr == "" {
		return nil, fmt.Errorf("registry-etcdhertz info addr missing port")
	}
	if host == "" || host == "::" {
		host = utils.LocalIP()
		if host == utils.UNKNOWN_IP_ADDR {
			return nil, fmt.Errorf("get local ip cwerror")
		}
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	if port <= 0 {
		return nil, cwerror.ErrMissPort
	}

	if info.Weight == 0 {
		info.Weight = registry.DefaultWeight
	}

	meta, err := sonic.Marshal(&internal.RegistryEntity{Weight: info.Weight, Tags: info.Tags})
	if err != nil {
		return nil, err
	}
	instanceKey := fmt.Sprintf("%s:%s", info.ServiceName, info.Addr.String())
	instance := &fargo.Instance{
		HostName:       instanceKey,
		InstanceId:     instanceKey,
		App:            info.ServiceName,
		IPAddr:         host,
		Port:           int(port),
		Status:         fargo.UP,
		DataCenterInfo: fargo.DataCenterInfo{Name: fargo.MyOwn},
	}

	instance.SetMetadataString(internal.Meta, string(meta))
	return instance, nil
}

func (e *eurekaRegistry) heartBeat(ctx context.Context, ins *fargo.Instance) {
	ticker := time.NewTicker(e.heatBeatInterval)
	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			if err := e.eurekaConn.HeartBeatInstance(ins); err != nil {
				hlog.Errorf("HERTZ: Heartbeat cwerror, reason: %s", err.Error())
			}
		}
	}
}
