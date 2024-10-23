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

package servicecomb

import (
	"os"
	"strconv"

	"github.com/cloudwego/kitex/pkg/klog"
)

const (
	SC_ENV_SERVER_ADDR     = "serverAddr"
	SC_ENV_PORT            = "serverPort"
	SC_DEFAULT_SERVER_ADDR = "127.0.0.1"
	SC_DEFAULT_PORT        = 30100
)

// SCPort Get ServiceComb port from environment variables
func SCPort() int64 {
	portText := os.Getenv(SC_ENV_PORT)
	if len(portText) == 0 {
		return SC_DEFAULT_PORT
	}
	port, err := strconv.ParseInt(portText, 10, 64)
	if err != nil {
		klog.Errorf("ParseInt failed,err:%s", err.Error())
		return SC_DEFAULT_PORT
	}
	return port
}

// SCAddr Get ServiceComb addr from environment variables
func SCAddr() string {
	addr := os.Getenv(SC_ENV_SERVER_ADDR)
	if len(addr) == 0 {
		return SC_DEFAULT_SERVER_ADDR
	}
	return addr
}
