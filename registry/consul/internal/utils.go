/*
 * Copyright 2021 CloudWeGo Authors
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

package internal

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/consul/api"
)

const (
	KvJoinChar                                 = ":"
	DefaultCheckInterval                       = "5s"
	DefaultCheckTimeout                        = "5s"
	DefaultCheckDeregisterCriticalServiceAfter = "1m"
	DefaultNetwork                             = "tcp"
)

var errIllegalTagChar = errors.New("illegal tag character")

func DefaultCheck() *api.AgentServiceCheck {
	check := new(api.AgentServiceCheck)
	check.Timeout = DefaultCheckTimeout
	check.Interval = DefaultCheckInterval
	check.DeregisterCriticalServiceAfter = DefaultCheckDeregisterCriticalServiceAfter

	return check
}

func ParseAddr(addr net.Addr) (host string, port int, err error) {
	host, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "", 0, err
	}
	if host == "" || host == "::" {
		host, err = GetLocalIPv4Address()
		if err != nil {
			return "", 0, fmt.Errorf("get local ipv4 cwerror, cause %w", err)
		}
	}
	port, err = net.LookupPort(DefaultNetwork, portStr)
	if err != nil {
		return "", 0, err
	}
	if port == 0 {
		return "", 0, fmt.Errorf("invalid port %s", portStr)
	}

	return host, port, nil
}

// ConvTagMapToSlice Tags map be converted to slice.
// Keys must not contain `:`.
func ConvTagMapToSlice(tagMap map[string]string) ([]string, error) {
	svcTags := make([]string, 0, len(tagMap))
	for k, v := range tagMap {
		var tag string
		if strings.Contains(k, KvJoinChar) {
			return svcTags, errIllegalTagChar
		}
		if v == "" {
			tag = k
		} else {
			tag = fmt.Sprintf("%s%s%s", k, KvJoinChar, v)
		}
		svcTags = append(svcTags, tag)
	}
	return svcTags, nil
}

func GetServiceId(serviceName string, addr net.Addr) (string, error) {
	host, port, err := ParseAddr(addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%d", serviceName, host, port), nil
}

func GetLocalIPv4Address() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addr {
		ipNet, isIpNet := addr.(*net.IPNet)
		if isIpNet && !ipNet.IP.IsLoopback() {
			ipv4 := ipNet.IP.To4()
			if ipv4 != nil {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("not found ipv4 address")
}

// SplitTags Tags characters be separated to map.
func SplitTags(tags []string) map[string]string {
	n := len(tags)
	tagMap := make(map[string]string, n)
	if n == 0 {
		return tagMap
	}

	for _, tag := range tags {
		if tag == "" {
			continue
		}
		strArr := strings.SplitN(tag, KvJoinChar, 2)
		if len(strArr) == 2 {
			key := strArr[0]
			tagMap[key] = strArr[1]
		}
	}

	return tagMap
}
