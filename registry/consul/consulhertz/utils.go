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

package consulhertz

import (
	"errors"
	"fmt"

	"github.com/cloudwego/hertz/pkg/app/server/registry"
	"github.com/cwgo-pkg/registry/consul/internal"
)

var errIllegalTagChar = errors.New("illegal tag character")

func getServiceId(info *registry.Info) (string, error) {
	host, port, err := internal.ParseAddr(info.Addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%d", info.ServiceName, host, port), nil
}
