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
	"testing"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"

	cwlogrus "github.com/cloudwego-contrib/cwgo-pkg/logging/logrus"

	"github.com/sirupsen/logrus"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logger := cwlogrus.NewLogger(cwlogrus.WithLogger(logrus.New()))

	logger.Logger().Info("log from origin otellogrus")

	logging.SetLogger(logger)
	logging.SetLevel(logging.LevelError)
	logging.SetLevel(logging.LevelWarn)
	logging.SetLevel(logging.LevelInfo)
	logging.SetLevel(logging.LevelDebug)
	logging.SetLevel(logging.LevelTrace)

	logging.Trace("trace")
	logging.Debug("debug")
	logging.Info("info")
	logging.Notice("notice")
	logging.Warn("warn")
	logging.Error("error")

	logging.Tracef("log level: %s", "trace")
	logging.Debugf("log level: %s", "debug")
	logging.Infof("log level: %s", "info")
	logging.Noticef("log level: %s", "notice")
	logging.Warnf("log level: %s", "warn")
	logging.Errorf("log level: %s", "error")

	logging.CtxTracef(ctx, "log level: %s", "trace")
	logging.CtxDebugf(ctx, "log level: %s", "debug")
	logging.CtxInfof(ctx, "log level: %s", "info")
	logging.CtxNoticef(ctx, "log level: %s", "notice")
	logging.CtxWarnf(ctx, "log level: %s", "warn")
	logging.CtxErrorf(ctx, "log level: %s", "error")
}
