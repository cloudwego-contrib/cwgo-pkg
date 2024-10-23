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

package degradation

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/bytedance/gopkg/lang/fastrand"
	"github.com/cloudwego/configmanager/iface"
	"github.com/cloudwego/kitex/pkg/acl"
)

var errRejected = errors.New("rejected by client degradation config")

var defaultDegradationConfig = &DegradationConfig{
	Enable:     false,
	Percentage: 0,
}

type DegradationConfig struct {
	Enable     bool `json:"enable"`
	Percentage int  `json:"percentage"`
}

// DeepCopy returns a copy of the current DegradationConfig
func (c *DegradationConfig) DeepCopy() iface.ConfigValueItem {
	result := &DegradationConfig{
		Enable:     c.Enable,
		Percentage: c.Percentage,
	}
	return result
}

// EqualsTo returns true if the current DegradationConfig equals to the other DegradationConfig
func (c *DegradationConfig) EqualsTo(other iface.ConfigValueItem) bool {
	o := other.(*DegradationConfig)
	return c.Enable == o.Enable && c.Percentage == o.Percentage
}

// DegradationContainer is a wrapper for DegradationConfig
type DegradationContainer struct {
	config atomic.Value
}

func NewDegradationContainer() *DegradationContainer {
	c := &DegradationContainer{}
	c.config.Store(defaultDegradationConfig)
	return c
}

// NotifyPolicyChange to receive policy when it changes
func (c *DegradationContainer) NotifyPolicyChange(cfg *DegradationConfig) {
	c.config.Store(cfg)
}

func (c *DegradationContainer) GetAclRule() acl.RejectFunc {
	return func(ctx context.Context, request interface{}) (reason error) {
		cfg := c.config.Load().(*DegradationConfig)
		if !cfg.Enable {
			return nil
		}
		if fastrand.Intn(100) < cfg.Percentage {
			return errRejected
		}
		return nil
	}
}
