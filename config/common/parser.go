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

package utils

import (
	"fmt"

	"github.com/bytedance/sonic"
	"sigs.k8s.io/yaml"
)

const (
	JSON ConfigType = "json"
	YAML ConfigType = "yaml"
	HCL  ConfigType = "hcl"
)

var _ ConfigParser = &parser{}

// ConfigParser the parser for Apollo config.
type ConfigParser interface {
	Decode(kind ConfigType, data string, config interface{}) error
}

type parser struct{}

// Decode decodes the data to struct in specified format.
func (p *parser) Decode(kind ConfigType, data string, config interface{}) error {
	switch kind {
	case JSON:
		return sonic.Unmarshal([]byte(data), config)
	case YAML:
		return yaml.Unmarshal([]byte(data), config)
	default:
		return fmt.Errorf("unsupported config data type %s", kind)
	}
}

// DefaultConfigParse default apollo config parser.
func DefaultConfigParse() ConfigParser {
	return &parser{}
}

// CustomFunction use for customize the config parameters.
type (
	ConfigType    string
	ConfigContent string
)

// ConfigParamConfig use for render the dataId or group info by go template, ref: https://pkg.go.dev/text/template
// The fixed key shows as below.
type ConfigParamConfig struct {
	Category          string
	ClientServiceName string
	ServerServiceName string
}
