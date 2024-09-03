// Copyright 2023 CloudWeGo Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package zookeeper

import "encoding/json"

const ( //`{{$Prefix}}/{{$ClientName}}/{{$ServerName}}/{{$ConfigCategory}}`
	ZookeeperDefaultServer     = "127.0.0.1:2181"
	ZookeeperDefaultClientPath = "{{.ClientServiceName}}/{{.ServerServiceName}}/{{.Category}}"
	ZookeeperDefaultServerPath = "{{.ServerServiceName}}/{{.Category}}"
	ZookeeperDefaultPrefix     = "/KitexConfig"
)

// CustomFunction use for customize the config parameters.
type CustomFunction func(*ConfigParam)

// ConfigParamConfig use for render the path info by go template, ref: https://pkg.go.dev/text/template
// The fixed key shows as below.
type ConfigParamConfig struct {
	Category          string
	ClientServiceName string
	ServerServiceName string
}

// ConfigParser the parser for zookeeper config.
type ConfigParser interface {
	Decode(data string, config interface{}) error
}
type parser struct{}

// Decode decodes the data to struct in specified format.
func (p *parser) Decode(data string, config interface{}) error {
	return json.Unmarshal([]byte(data), config)
}

// DefaultConfigParser default zookeeper config parser.
func defaultConfigParser() ConfigParser {
	return &parser{}
}
