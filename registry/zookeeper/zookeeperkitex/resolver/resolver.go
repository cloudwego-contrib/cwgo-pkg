/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package resolver

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper/common"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper/zookeeperkitex/entity"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/go-zookeeper/zk"
)

type zookeeperResolver struct {
	conn *zk.Conn
}

// NewZookeeperResolver create a zookeeper based resolver
func NewZookeeperResolver(servers []string, sessionTimeout time.Duration) (discovery.Resolver, error) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	return &zookeeperResolver{conn: conn}, nil
}

// NewZookeeperResolver create a zookeeper based resolver with auth
func NewZookeeperResolverWithAuth(servers []string, sessionTimeout time.Duration, user, password string) (discovery.Resolver, error) {
	conn, _, err := zk.Connect(servers, sessionTimeout)
	if err != nil {
		return nil, err
	}
	auth := []byte(fmt.Sprintf("%s:%s", user, password))
	err = conn.AddAuth(common.Scheme, auth)
	if err != nil {
		return nil, err
	}
	return &zookeeperResolver{conn: conn}, nil
}

// Target implements the Resolver interface.
func (r *zookeeperResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) string {
	return target.ServiceName()
}

// Resolve implements the Resolver interface.
func (r *zookeeperResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	path := desc
	if !strings.HasPrefix(path, common.Separator) {
		path = common.Separator + path
	}
	eps, err := r.getEndPoints(path)
	if err != nil {
		return discovery.Result{}, err
	}
	if len(eps) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	instances, err := r.getInstances(eps, path)
	if err != nil {
		return discovery.Result{}, err
	}
	res := discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: instances,
	}
	return res, nil
}

func (r *zookeeperResolver) getEndPoints(path string) ([]string, error) {
	child, _, err := r.conn.Children(path)
	return child, err
}

func (r *zookeeperResolver) detailEndPoints(path, ep string) (discovery.Instance, error) {
	data, _, err := r.conn.Get(path + common.Separator + ep)
	if err != nil {
		return nil, err
	}
	en := new(entity.RegistryEntity)
	err = json.Unmarshal(data, en)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data [%s] cwerror, cause %w", data, err)
	}
	return discovery.NewInstance("tcp", ep, en.Weight, en.Tags), nil
}

func (r *zookeeperResolver) getInstances(eps []string, path string) ([]discovery.Instance, error) {
	instances := make([]discovery.Instance, 0, len(eps))
	for _, ep := range eps {
		if host, port, err := net.SplitHostPort(ep); err == nil {
			if port == "" {
				return []discovery.Instance{}, fmt.Errorf("missing port when parse node [%s]", ep)
			}
			if host == "" {
				return []discovery.Instance{}, fmt.Errorf("missing host when parse node [%s]", ep)
			}
			ins, err := r.detailEndPoints(path, ep)
			if err != nil {
				return []discovery.Instance{}, fmt.Errorf("detail endpoint [%s] info cwerror, cause %w", ep, err)
			}
			instances = append(instances, ins)
		} else {
			return []discovery.Instance{}, fmt.Errorf("parse node [%s] cwerror, details info [%w]", ep, err)
		}
	}
	return instances, nil
}

// Diff implements the Resolver interface.
func (r *zookeeperResolver) Diff(key string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(key, prev, next)
}

// Name implements the Resolver interface.
func (r *zookeeperResolver) Name() string {
	return "zookeeper"
}
