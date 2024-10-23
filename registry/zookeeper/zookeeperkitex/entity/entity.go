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

package entity

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper/common"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/zookeeper/zookeeperkitex/utils"
	"github.com/cloudwego/kitex/pkg/registry"
)

type RegistryEntity struct {
	Weight int
	Tags   map[string]string
}

type NodeEntity struct {
	*registry.Info
}

func (n *NodeEntity) Path() (string, error) {
	return buildPath(n.Info)
}

func (n *NodeEntity) Content() ([]byte, error) {
	return json.Marshal(&RegistryEntity{Weight: n.Weight, Tags: n.Tags})
}

func MustNewNodeEntity(ri *registry.Info) *NodeEntity {
	return &NodeEntity{ri}
}

// path format as follows:
// /{serviceName}/{ip}:{port}
func buildPath(info *registry.Info) (string, error) {
	var path string
	if info == nil {
		return "", fmt.Errorf("registry info can't be nil")
	}
	if info.ServiceName == "" {
		return "", fmt.Errorf("registry info service name can't be empty")
	}
	if info.Addr == nil {
		return "", fmt.Errorf("registry info addr can't be nil")
	}
	if !strings.HasPrefix(info.ServiceName, common.Separator) {
		path = common.Separator + info.ServiceName
	}

	if host, port, err := net.SplitHostPort(info.Addr.String()); err == nil {
		if port == "" {
			return "", fmt.Errorf("registry info addr missing port")
		}
		ip := net.ParseIP(host)
		if ip == nil || ip.IsUnspecified() {
			ipv4, err := utils.GetLocalIPv4Address()
			if err != nil {
				return "", fmt.Errorf("get local ipv4 cwerror, cause %w", err)
			}
			path = path + common.Separator + ipv4 + ":" + port
		} else {
			path = path + common.Separator + host + ":" + port
		}
	} else {
		return "", fmt.Errorf("parse registry info addr cwerror")
	}
	return path, nil
}
