/*
 * Copyright 2024 CloudWeGo Authors
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

package logrus_test

import (
	"context"
	"testing"

	cwlogrus "github.com/cloudwego-contrib/cwgo-pkg/log/logging/logrus"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/sirupsen/logrus"
)

func TestKLogger(t *testing.T) {
	ctx := context.Background()

	logger := cwlogrus.NewKLogger(cwlogrus.WithLogger(logrus.New()))

	logger.Info("log from origin otellogrus")

	klog.SetLogger(logger)
	klog.SetLevel(klog.LevelError)
	klog.SetLevel(klog.LevelWarn)
	klog.SetLevel(klog.LevelInfo)
	klog.SetLevel(klog.LevelDebug)
	klog.SetLevel(klog.LevelTrace)

	klog.Trace("trace")
	klog.Debug("debug")
	klog.Info("info")
	klog.Notice("notice")
	klog.Warn("warn")
	klog.Error("error")

	klog.Tracef("log level: %s", "trace")
	klog.Debugf("log level: %s", "debug")
	klog.Infof("log level: %s", "info")
	klog.Noticef("log level: %s", "notice")
	klog.Warnf("log level: %s", "warn")
	klog.Errorf("log level: %s", "error")

	klog.CtxTracef(ctx, "log level: %s", "trace")
	klog.CtxDebugf(ctx, "log level: %s", "debug")
	klog.CtxInfof(ctx, "log level: %s", "info")
	klog.CtxNoticef(ctx, "log level: %s", "notice")
	klog.CtxWarnf(ctx, "log level: %s", "warn")
	klog.CtxErrorf(ctx, "log level: %s", "error")
}
