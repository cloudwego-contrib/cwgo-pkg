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

package test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/cwerror"
	registry2 "github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/eurekakitex/registry"
	"github.com/cloudwego-contrib/cwgo-pkg/registry/eureka/eurekakitex/resolver"

	"github.com/cloudwego/kitex/pkg/kerrors"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"

	"github.com/stretchr/testify/assert"
)

func TestEurekaDiscovery(t *testing.T) {
	var err error
	r := registry2.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)
	tags := map[string]string{"idc": "hl"}
	addr := &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8888}
	info := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         tags,
		Addr:         addr,
	}
	// 1. register one instance
	err = r.Register(info)
	assert.Nil(t, err)

	res := resolver.NewEurekaResolver([]string{"http://127.0.0.1:8761/eureka"})
	target := res.Target(context.Background(), rpcinfo.NewEndpointInfo("test", "", nil, nil))

	// 2. expect return one instance when do discovery
	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Instances))

	instance := result.Instances[0]
	assert.Equal(t, addr.String(), instance.Address().String())
	assert.Equal(t, info.Weight, instance.Weight())

	for k, v := range info.Tags {
		v1, exist := instance.Tag(k)
		assert.Equal(t, true, exist)
		assert.Equal(t, v, v1)
	}

	// 3. deregister one instance that register instantly
	err = r.Deregister(info)
	assert.Nil(t, err)

	// 4. expect no instance when do discovery again
	result, err = res.Resolve(context.Background(), target)
	assert.Equal(t, kerrors.ErrNoMoreInstance, err)
	assert.Equal(t, 0, len(result.Instances))
}

func TestEurekaDiscoveryWithMultipleInstance(t *testing.T) {
	r := registry2.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)
	info1 := &registry.Info{
		ServiceName:  "test",
		Weight:       11,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8888},
	}
	info2 := &registry.Info{
		ServiceName:  "test",
		Weight:       12,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8889},
	}
	info3 := &registry.Info{
		ServiceName:  "test",
		Weight:       13,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8890},
	}
	addrMap := map[string]int{
		info1.Addr.String(): info1.Weight,
		info2.Addr.String(): info2.Weight,
		info3.Addr.String(): info3.Weight,
	}

	// 1. register three instance
	assert.Nil(t, r.Register(info1))
	assert.Nil(t, r.Register(info2))
	assert.Nil(t, r.Register(info3))

	res := resolver.NewEurekaResolver([]string{"http://127.0.0.1:8761/eureka"})

	target := res.Target(context.Background(), rpcinfo.NewEndpointInfo("test", "", nil, nil))

	// 2. expect three instances when do discovery
	result, err := res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Len(t, result.Instances, 3)
	instances := result.Instances
	for _, instance := range instances {
		addr := instance.Address().String()
		weight, ok := addrMap[addr]
		assert.Equal(t, true, ok)
		assert.Equal(t, weight, instance.Weight())
		v1, exist := instance.Tag("idc")
		assert.Equal(t, true, exist)
		assert.Equal(t, "hl", v1)
	}

	// 3. deregister two instances that register instantly
	assert.Nil(t, r.Deregister(info1))
	assert.Nil(t, r.Deregister(info2))

	// 4. expect just one instances when do discovery
	result, err = res.Resolve(context.Background(), target)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(result.Instances))
	instance := result.Instances[0]
	assert.Equal(t, info3.Weight, instance.Weight())
	assert.Equal(t, info3.Addr.String(), instance.Address().String())

	// 5. deregister finally one instance that register instantly
	assert.Nil(t, r.Deregister(info3))

	// 6. expect no instances when do discovery
	result, err = res.Resolve(context.Background(), target)
	assert.Equal(t, kerrors.ErrNoMoreInstance, err)
	assert.Equal(t, 0, len(result.Instances))
}

func TestEurekaDiscoveryWithInvalidInstanceInfo(t *testing.T) {
	r := registry2.NewEurekaRegistry([]string{"http://127.0.0.1:8761/eureka"}, 11*time.Second)

	// 1. try to register nil instance
	assert.Equal(t, cwerror.ErrNilInfo, r.Register(nil))

	info1 := &registry.Info{
		ServiceName:  "",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8888},
	}
	// 2. try to register one instance that not have serviceName
	assert.Equal(t, cwerror.ErrEmptyServiceName, r.Register(info1))

	info2 := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         nil,
	}
	// 3. try to register one instance that not have addr
	assert.Equal(t, cwerror.ErrNilAddr, r.Register(info2))

	info3 := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{Port: 8889},
	}
	// 4. try to register one instance that not have ip
	assert.Equal(t, cwerror.ErrMissIP, r.Register(info3))

	info4 := &registry.Info{
		ServiceName:  "test",
		Weight:       10,
		PayloadCodec: "thrift",
		Tags:         map[string]string{"idc": "hl"},
		Addr:         &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1)},
	}
	// 5. try to register one instance that not have port
	assert.Equal(t, cwerror.ErrMissPort, r.Register(info4))
}
