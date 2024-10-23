// Copyright 2024 CloudWeGo Authors.
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
	"net"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/options"

	nacosregistry "github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/nacoskitex/v2/registry"
	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/assert"
)

var (
	nacosCli naming_client.INamingClient
	svcName  = "demo.kitex-contrib.local"
	svcAddr  = net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8080}
	svcInfo  = &registry.Info{
		ServiceName: svcName,
		Addr:        &svcAddr,
		Weight:      999,
		StartTime:   time.Now(),
		Tags:        map[string]string{"env": "local"},
	}
)

func init() {
	cli, err := getNacosClient()
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	err = nacosregistry.NewNacosRegistry(cli).Register(svcInfo)
	if err != nil {
		return
	}
	time.Sleep(time.Second)
	nacosCli = cli
}

func getNacosClient() (naming_client.INamingClient, error) {
	// create ServerConfig
	sc := []constant.ServerConfig{
		*constant.NewServerConfig("127.0.0.1", 8848, constant.WithContextPath("/nacos"), constant.WithScheme("http")),
	}

	// create ClientConfig
	cc := *constant.NewClientConfig(
		constant.WithTimeoutMs(50000),
		constant.WithUpdateCacheWhenEmpty(true),
		constant.WithNotLoadCacheAtStart(true),
	)

	// create naming client
	newClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	return newClient, err
}

// TestNacosResolverResolve test Resolve a service
func TestNacosResolverResolve(t *testing.T) {
	type fields struct {
		cli naming_client.INamingClient
	}
	type args struct {
		ctx  context.Context
		desc string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    discovery.Result
		wantErr bool
	}{
		{
			name: "common",
			args: args{
				ctx:  context.Background(),
				desc: svcName,
			},
			fields: fields{cli: nacosCli},
		},
		{
			name: "wrong desc",
			args: args{
				ctx:  context.Background(),
				desc: "xxxx.kitex-contrib.local",
			},
			fields:  fields{cli: nacosCli},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewNacosResolver(tt.fields.cli)
			_, err := n.Resolve(tt.args.ctx, tt.args.desc)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() cwerror = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && !strings.Contains(err.Error(), "instance list is empty") {
				t.Errorf("Resolve err is not expectant")
				return
			}
		})
	}

	err := nacosregistry.NewNacosRegistry(nacosCli).Deregister(svcInfo)
	assert.Nil(t, err)
}

// TestNacosResolverDifferentCluster test NewNacosResolver WithCluster option
func TestNacosResolverDifferentCluster(t *testing.T) {
	ctx := context.Background()
	n := NewNacosResolver(nacosCli)
	got, err := n.Resolve(ctx, svcName)
	assert.Nil(t, err)
	assert.NotNil(t, got)
	assert.Equal(t, svcName, got.CacheKey)
	if assert.NotEmpty(t, got.Instances) {
		gotSvc := got.Instances[0]
		assert.Equal(t, gotSvc.Address().String(), svcAddr.String())
	}

	n = NewNacosResolver(nacosCli, options.WithResolverCluster("OTHER"))
	_, err = n.Resolve(ctx, svcName)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "instance list is empty")
}

// TestNewDefaultNacosResolver test new a default nacos resolver
func TestNewDefaultNacosResolver(t *testing.T) {
	r, err := NewDefaultNacosResolver()
	assert.Nil(t, err)
	assert.NotNil(t, r)
}
