// Copyright 2024 CloudWeGo Authors
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

package parser

import (
	"fmt"

	"github.com/bytedance/sonic"
	"sigs.k8s.io/yaml"
)

type ConfigParam struct {
	Type ConfigType
}

type ConfigType string

const (
	JSON ConfigType = "json"
	YAML ConfigType = "yaml"
)

// ConfigParser the parser for config file.
type ConfigParser interface {
	Decode(kind ConfigType, data []byte, config interface{}) error
}

type Parser struct{}

type ConfigManager interface {
	GetConfig(key string) interface{}
}

var _ ConfigParser = &Parser{}

// Decode decodes the data to struct in specified format.
func (p *Parser) Decode(kind ConfigType, data []byte, config interface{}) error {
	switch kind {
	case JSON:
		return sonic.Unmarshal(data, config)
	case YAML:
		return yaml.Unmarshal(data, config)
	default:
		return fmt.Errorf("unsupported config data type %s", kind)
	}
}

func DefaultConfigParser() ConfigParser {
	return &Parser{}
}

func DefaultConfigParam() *ConfigParam {
	return &ConfigParam{
		Type: JSON,
	}
}
