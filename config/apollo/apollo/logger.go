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

package apollo

import (
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/cloudwego/kitex/pkg/klog"
)

func init() {
	log.InitLogger(NewCustomApolloLogger())
}

type customApolloLogger struct{}

func NewCustomApolloLogger() log.LoggerInterface {
	return customApolloLogger{}
}

func (m customApolloLogger) Info(v ...interface{}) {
	klog.Info(v...)
}

func (m customApolloLogger) Warn(v ...interface{}) {
	klog.Warn(v...)
}

func (m customApolloLogger) Error(v ...interface{}) {
	klog.Error(v...)
}

func (m customApolloLogger) Debug(v ...interface{}) {
	klog.Debug(v)
}

func (m customApolloLogger) Infof(fmt string, v ...interface{}) {
	klog.Infof(fmt, v...)
}

func (m customApolloLogger) Warnf(fmt string, v ...interface{}) {
	klog.Warnf(fmt, v...)
}

func (m customApolloLogger) Errorf(fmt string, v ...interface{}) {
	klog.Errorf(fmt, v...)
}

func (m customApolloLogger) Debugf(fmt string, v ...interface{}) {
	klog.Debugf(fmt, v...)
}
