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

package nacos

import (
	"testing"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/nacos/common"

	"github.com/stretchr/testify/assert"
)

// TestEnvFunc test env func
func TestEnvFunc(t *testing.T) {
	assert.Equal(t, int64(8848), NacosPort())
	assert.Equal(t, "127.0.0.1", common.NacosAddr())
}

func TestParseTags(t *testing.T) {
	assert.Equal(t, parseTags("k1=v1,k2=v2"), map[string]string{
		"cloudwego.nacos.client": "kitex",
		"k1":                     "v1",
		"k2":                     "v2",
	})
}
