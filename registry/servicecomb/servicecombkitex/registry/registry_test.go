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
	"github.com/cloudwego-contrib/cwgo-pkg/registry/servicecomb/options"
	"net"
	"testing"
	"time"

	"github.com/go-chassis/sc-client"

	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/stretchr/testify/assert"
)

const (
	ServiceName   = "demo.kitex-contrib.local"
	AppId         = "DEFAULT"
	Version       = "1.0.0"
	LatestVersion = "latest"
	HostName      = "DEFAULT"
)

func getSCClient() (*sc.Client, error) {
	client, err := sc.NewClient(sc.Options{
		Endpoints: []string{"127.0.0.1:30100"},
	})
	if err != nil {
		return nil, err
	}
	return client, nil
}

func TestNewDefaultSCRegistry(t *testing.T) {
	client, err := getSCClient()
	if err != nil {
		t.Errorf("err:%v", err)
	}
	got := NewSCRegistry(client, options.WithAppId(AppId), options.WithVersionRule(Version))
	assert.NotNil(t, got)
}

// test registry a service
func TestSCRegistryRegister(t *testing.T) {
	client, err := getSCClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	type fields struct {
		cli *sc.Client
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
				ServiceName: ServiceName,
				Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 3000},
			}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSCRegistry(tt.fields.cli, options.WithAppId(AppId), options.WithVersionRule(Version), options.WithHostName(HostName))
			if err := n.Register(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Register() cwerror = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// test deregister a service
func TestSCRegistryDeregister(t *testing.T) {
	client, err := getSCClient()
	if err != nil {
		t.Errorf("err:%v", err)
		return
	}
	type fields struct {
		cli *sc.Client
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
				ServiceName: ServiceName,
			}},
			fields:  fields{client},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSCRegistry(tt.fields.cli, options.WithAppId(AppId), options.WithVersionRule(Version), options.WithHostName(HostName))
			if err := n.Deregister(tt.args.info); (err != nil) != tt.wantErr {
				t.Errorf("Deregister() cwerror = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// test register several instances and deregister
func TestSCMultipleInstances(t *testing.T) {
	client, err := getSCClient()
	assert.Nil(t, err)
	time.Sleep(time.Second)
	got := NewSCRegistry(client, options.WithAppId(AppId), options.WithVersionRule(Version), options.WithHostName(HostName), options.WithHeartbeatInterval(5))
	if !assert.NotNil(t, got) {
		t.Errorf("err: new registry fail")
		return
	}
	time.Sleep(time.Second)

	err = got.Register(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8081},
	})
	assert.Nil(t, err)
	err = got.Register(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8082},
	})
	assert.Nil(t, err)
	err = got.Register(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
	})
	assert.Nil(t, err)

	time.Sleep(time.Second)
	err = got.Deregister(&registry.Info{
		ServiceName: ServiceName,
		Addr:        &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8083},
	})
	assert.Nil(t, err)
	_, err = client.FindMicroServiceInstances("", AppId, ServiceName, LatestVersion, sc.WithoutRevision())
	assert.Nil(t, err)
}
