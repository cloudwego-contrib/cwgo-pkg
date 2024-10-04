// Copyright 2022 CloudWeGo Authors.
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

package otelzap

import (
	"testing"

	cwzap "github.com/cloudwego-contrib/cwgo-pkg/logging/zap"
	"github.com/stretchr/testify/assert"
)

func TestWithLogger(t *testing.T) {
	l := NewLogger(WithLogger(cwzap.NewLogger()))
	for _, v := range extraKeys {
		assert.Contains(t, l.GetExtraKeys(), v)
	}
}
