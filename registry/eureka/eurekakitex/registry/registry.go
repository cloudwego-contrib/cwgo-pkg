// Copyright 2021 CloudWeGo authors.
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
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/cwerror"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/internal"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/hudl/fargo"
)

type eurekaHeartbeat struct {
	cancel      context.CancelFunc
	instanceKey string
}

type eurekaRegistry struct {
	eurekaConn       *fargo.EurekaConnection
	heatBeatInterval time.Duration
	lock             *sync.RWMutex
	registryIns      map[string]*eurekaHeartbeat
}

// NewEurekaRegistry creates a eureka registry.
func NewEurekaRegistry(servers []string, heatBeatInterval time.Duration) registry.Registry {
	conn := fargo.NewConn(servers...)
	return &eurekaRegistry{
		eurekaConn:       &conn,
		heatBeatInterval: heatBeatInterval,
		registryIns:      make(map[string]*eurekaHeartbeat),
		lock:             &sync.RWMutex{},
	}
}

// Register register a server with given registry info.
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

// Deregister deregister a server with given registry info.
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
	if host == "" || host == "::" {
		return nil, cwerror.ErrMissIP
	}

	port, err := strconv.ParseInt(portStr, 10, 64)
	if err != nil {
		return nil, err
	}
	if port <= 0 {
		return nil, cwerror.ErrMissPort
	}

	if info.Weight == 0 {
		info.Weight = discovery.DefaultWeight
	}

	meta, err := json.Marshal(&internal.RegistryEntity{Weight: info.Weight, Tags: info.Tags})
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
				klog.Errorf("heartBeat cwerror,err=%+v", err)
			}
		}
	}
}
