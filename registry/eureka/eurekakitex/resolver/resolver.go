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

package resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/internal"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/hudl/fargo"
)

// eurekaResolver is a resolver using eureka.
type eurekaResolver struct {
	eurekaConn *fargo.EurekaConnection
}

// NewEurekaResolver creates a eureka resolver.
func NewEurekaResolver(servers []string) discovery.Resolver {
	conn := fargo.NewConn(servers...)
	return &eurekaResolver{eurekaConn: &conn}
}

// Target implements the Resolver interface.
func (r *eurekaResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) string {
	return target.ServiceName()
}

// Resolve implements the Resolver interface.
func (r *eurekaResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	application, err := r.eurekaConn.GetApp(desc)
	if err != nil {
		if errors.As(err, &fargo.AppNotFoundError{}) {
			return discovery.Result{}, kerrors.ErrNoMoreInstance
		}
		return discovery.Result{}, err
	}

	eurekaInstances := application.Instances
	instances, err := r.instances(eurekaInstances)
	if err != nil {
		return discovery.Result{}, err
	}

	return discovery.Result{CacheKey: desc, Cacheable: true, Instances: instances}, nil
}

// Diff implements the Resolver interface.
func (r *eurekaResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name implements the Resolver interface.
func (r *eurekaResolver) Name() string {
	return internal.Eureka
}

func (r *eurekaResolver) instances(instances []*fargo.Instance) ([]discovery.Instance, error) {
	res := make([]discovery.Instance, 0, len(instances))
	for _, instance := range instances {
		dInstance, err := r.instance(instance)
		if err != nil {
			return nil, err
		}
		res = append(res, dInstance)
	}

	return res, nil
}

func (r *eurekaResolver) instance(instance *fargo.Instance) (discovery.Instance, error) {
	var dInstance discovery.Instance
	var e internal.RegistryEntity
	meta, err := instance.Metadata.GetString(internal.Meta)
	if err != nil {
		return dInstance, err
	}
	if err = json.Unmarshal([]byte(meta), &e); err != nil {
		return dInstance, err
	}

	return discovery.NewInstance(internal.TCP, fmt.Sprintf("%s:%d", instance.IPAddr, instance.Port), e.Weight, e.Tags), nil
}
