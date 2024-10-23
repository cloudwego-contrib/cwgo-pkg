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

import "github.com/cloudwego/kitex/pkg/limiter"

// ServerFileConfig is config of a service
type ServerFileConfig struct {
	Limit limiter.LimiterConfig `mapstructure:"limit"`
}

// ServerFileManager is a map of service names to ServerFileConfig
type ServerFileManager map[string]*ServerFileConfig

// GetConfig returns the config from Manager by key
func (s *ServerFileManager) GetConfig(key string) interface{} {
	config, exist := (*s)[key]

	if !exist {
		return nil
	}

	return config
}
