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

package consul

import (
	"fmt"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/consul/internal"
	"github.com/cloudwego/kitex/pkg/registry"
)

func getServiceId(info *registry.Info) (string, error) {
	host, port, err := internal.ParseAddr(info.Addr)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%s:%d", info.ServiceName, host, port), nil
}
