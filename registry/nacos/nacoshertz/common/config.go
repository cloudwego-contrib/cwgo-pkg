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

package common

import (
	"os"
	"strconv"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/common"

	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

// NewDefaultNacosConfig create a default Nacos client.
func NewDefaultNacosConfig() (naming_client.INamingClient, error) {
	sc := []constant.ServerConfig{
		*constant.NewServerConfig(common.NacosAddr(), uint64(NacosPort())),
	}
	cc := constant.ClientConfig{
		NamespaceId:         common.NacosNameSpaceID(),
		RegionId:            common.NacosDefaultRegionID,
		CustomLogger:        NewCustomNacosLogger(),
		NotLoadCacheAtStart: true,
	}
	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &cc,
			ServerConfigs: sc,
		},
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// NacosPort Get Nacos port from environment variables.
func NacosPort() int64 {
	portText := os.Getenv(common.NacosEnvServerPort)
	if len(portText) == 0 {
		return common.NacosDefaultPort
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		hlog.Errorf("ParseInt failed,err:%s", err.Error())
		return common.NacosDefaultPort
	}
	return port
}
