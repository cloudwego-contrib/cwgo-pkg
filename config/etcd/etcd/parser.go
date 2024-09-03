// Copyright 2023 CloudWeGo Authors
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

package etcd

import (
	"encoding/json"
	"time"
)

const (
	EtcdDefaultNode         = "http://127.0.0.1:2379"
	EtcdDefaultConfigPrefix = "/KitexConfig"
	EtcdDefaultTimeout      = 5 * time.Second
	EtcdDefaultClientPath   = "{{.ClientServiceName}}/{{.ServerServiceName}}/{{.Category}}"
	EtcdDefaultServerPath   = "{{.ServerServiceName}}/{{.Category}}"
)

var _ ConfigParser = &parser{}

// CustomFunction use for customize the config parameters.
type CustomFunction func(*Key)

// ConfigParamConfig use for render the path or prefix info by go template, ref: https://pkg.go.dev/text/template
// The fixed key shows as below.
type ConfigParamConfig struct {
	Category          string
	ClientServiceName string
	ServerServiceName string
}

// ConfigParser the parser for etcd config.
type ConfigParser interface {
	Decode(data string, config interface{}) error
}

type parser struct{}

// Decode decodes the data to struct in specified format.
func (p *parser) Decode(data string, config interface{}) error {
	return json.Unmarshal([]byte(data), config)
}

// DefaultConfigParse default etcd config parser.
func defaultConfigParse() ConfigParser {
	return &parser{}
}
