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
	"net"
	"testing"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/nacos"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stretchr/testify/assert"
)

func getNacosClient() (naming_client.INamingClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848),
	}

	cc := constant.ClientConfig{
		NamespaceId:         "public",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		CustomLogger:        nacos.NewCustomNacosLogger(),
		CacheDir:            "/tmp/nacos/cache",
	}

	return clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
}

// TestNewNacosRegistry test new a nacos registry
func TestNewNacosRegistry(t *testing.T) {
	client, err := getNacosClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	got := NewNacosRegistry(client, options.WithCluster("DEFAULT"), options.WithGroup("DEFAULT_GROUP"))
	assert.NotNil(t, got)
}

// TestNewNacosRegistry test registry a service
func TestNacosRegistryRegister(t *testing.T) {
	client, err := getNacosClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	type fields struct {
		cli naming_client.INamingClient
	}
	type args struct {
		info *registry.Info
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "common",
			fields: fields{client},
			args: args{info: &registry.Info{
				ServiceName: "demo.kitex-contrib.local",
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080},
				Weight:      999,
				StartTime:   time.Now(),
				Tags:        map[string]string{"env": "local"},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNacosRegistry(tt.fields.cli, options.WithCluster("DEFAULT"), options.WithGroup("DEFAULT_GROUP"))
			if err := n.Register(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Register() cwerror = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNacosRegistryDeregister test deregister a service
func TestNacosRegistryDeregister(t *testing.T) {
	client, err := getNacosClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	type fields struct {
		cli naming_client.INamingClient
	}
	type args struct {
		info *registry.Info
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "common",
			args: args{info: &registry.Info{
				ServiceName: "demo.kitex-contrib.local",
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080},
				Weight:      999,
				StartTime:   time.Now(),
				Tags:        map[string]string{"env": "local"},
			}},
			fields:  fields{client},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNacosRegistry(tt.fields.cli, options.WithCluster("DEFAULT"), options.WithGroup("DEFAULT_GROUP"))
			if err := n.Deregister(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Deregister() cwerror = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestNacosMultipleInstances test registry multiple service,then deregister one
func TestNacosMultipleInstances(t *testing.T) {
	var (
		svcName     = "MultipleInstances_V1"
		clusterName = "TheCluster"
		groupName   = "TheGroup"
	)
	client, err := getNacosClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	time.Sleep(time.Second)
	got := NewNacosRegistry(client, options.WithCluster(clusterName), options.WithGroup(groupName))
	if !assert.NotNil(t, got) {
		t.Errorf("err: new registry fail")
		return
	}
	time.Sleep(time.Second)
	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8081},
	})
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8082},
	})
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
	})
	assert.Nil(t, err)

	time.Sleep(time.Second)
	res, err := client.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res), "successful register not three")

	time.Sleep(time.Second)
	err = got.Deregister(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
	})
	assert.Nil(t, err)

	time.Sleep(time.Second * 3)
	res, err = client.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(res), "deregister one, instances num should be two") {
		for _, i := range res {
			assert.Equal(t, "127.0.0.1", i.Ip)
			assert.Contains(t, []uint64{8081, 8082}, i.Port)
		}
	}
}

// TestNewDefaultNacosRegistry test new a default nacos registry
func TestNewDefaultNacosRegistry(t *testing.T) {
	r, err := NewDefaultNacosRegistry()
	assert.Nil(t, err)
	assert.NotNil(t, r)
}

// TestNacosMultipleInstancesWithDefaultNacosRegistry use DefaultNacosRegistry to test registry multiple service,then deregister one
func TestNacosMultipleInstancesWithDefaultNacosRegistry(t *testing.T) {
	var (
		svcName     = "MultipleInstances_V1"
		clusterName = "TheCluster"
		groupName   = "TheGroup"
	)
	got, err := NewDefaultNacosRegistry(options.WithCluster(clusterName), options.WithGroup(groupName))
	assert.Nil(t, err)
	if !assert.NotNil(t, got) {
		t.Errorf("err: new registry fail")
		return
	}
	time.Sleep(time.Second)
	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8081},
	})
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8082},
	})
	assert.Nil(t, err)
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}

	err = got.Register(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
	})
	assert.Nil(t, err)

	time.Sleep(time.Second * 1)
	client, err := getNacosClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	res, err := client.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, len(res), "successful register not three")

	time.Sleep(time.Second)
	err = got.Deregister(&registry.Info{
		ServiceName: svcName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
	})
	assert.Nil(t, err)

	time.Sleep(time.Second * 3)
	res, err = client.SelectAllInstances(vo.SelectAllInstancesParam{
		ServiceName: svcName,
		GroupName:   groupName,
		Clusters:    []string{clusterName},
	})
	assert.Nil(t, err)
	if assert.Equal(t, 2, len(res), "deregister one, instances num should be two") {
		for _, i := range res {
			assert.Equal(t, "127.0.0.1", i.Ip)
			assert.Contains(t, []uint64{8081, 8082}, i.Port)
		}
	}
}

func TestMergeTags(t *testing.T) {
	t1 := map[string]string{
		"k1": "v1",
		"k2": "v2",
	}
	t2 := map[string]string{
		"k3": "v3",
		"k4": "v4",
	}
	merged := mergeTags(t1, t2)
	assert.Equal(t, merged, map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
		"k4": "v4",
	})
	assert.Equal(t, t1, map[string]string{
		"k1": "v1",
		"k2": "v2",
	})
	assert.Equal(t, t2, map[string]string{
		"k3": "v3",
		"k4": "v4",
	})
}
