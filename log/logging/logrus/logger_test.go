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

package logrus_test

import (
	"context"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"testing"

	cwlogrus "github.com/cloudwego-contrib/cwgo-pkg/log/logging/logrus"

	"github.com/sirupsen/logrus"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logger := cwlogrus.NewLogger(cwlogrus.WithLogger(logrus.New()))

	logger.Logger().Info("log from origin otellogrus")

	hlog.SetLogger(logger)
	hlog.SetLevel(hlog.LevelError)
	hlog.SetLevel(hlog.LevelWarn)
	hlog.SetLevel(hlog.LevelInfo)
	hlog.SetLevel(hlog.LevelDebug)
	hlog.SetLevel(hlog.LevelTrace)

	hlog.Trace("trace")
	hlog.Debug("debug")
	hlog.Info("info")
	hlog.Notice("notice")
	hlog.Warn("warn")
	hlog.Error("error")

	hlog.Tracef("log level: %s", "trace")
	hlog.Debugf("log level: %s", "debug")
	hlog.Infof("log level: %s", "info")
	hlog.Noticef("log level: %s", "notice")
	hlog.Warnf("log level: %s", "warn")
	hlog.Errorf("log level: %s", "error")

	hlog.CtxTracef(ctx, "log level: %s", "trace")
	hlog.CtxDebugf(ctx, "log level: %s", "debug")
	hlog.CtxInfof(ctx, "log level: %s", "info")
	hlog.CtxNoticef(ctx, "log level: %s", "notice")
	hlog.CtxWarnf(ctx, "log level: %s", "warn")
	hlog.CtxErrorf(ctx, "log level: %s", "error")
}
