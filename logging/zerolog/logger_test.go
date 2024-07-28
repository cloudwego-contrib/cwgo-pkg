// Copyright 2024 CloudWeGo Authors.
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

package zerolog

import (
	"context"
	"github.com/cloudwego-contrib/obs-opentelemetry/logging"
	"go.opentelemetry.io/otel"
	"os"
	"testing"

	"github.com/rs/zerolog"
)

func TestLogger(t *testing.T) {
	ctx := context.Background()

	logger := NewLogger(
		WithTraceErrorSpanLevel(zerolog.WarnLevel),
		WithRecordStackTraceInSpan(true),
	)

	logger.Logger().Info().Msg("log from origin zerolog")

	logging.SetLogger(logger)

	logging.SetLevel(logging.LevelDebug)

	logging.SetOutput(os.Stderr)

	tracer := otel.Tracer("test otel std logger")

	ctx, span := tracer.Start(ctx, "root")

	logging.CtxInfof(ctx, "hello %s", "world")

	defer span.End()

	ctx, child := tracer.Start(ctx, "child")

	logging.CtxDebugf(ctx, "foo %s", "bar")
	logging.CtxTracef(ctx, "foo %s", "bar")
	logging.CtxInfof(ctx, "foo %s", "bar")
	logging.CtxNoticef(ctx, "foo %s", "bar")
	logging.CtxWarnf(ctx, "foo %s", "bar")
	logging.CtxErrorf(ctx, "foo %s", "bar")
	logging.Debugf("foo %s", "bar")
	logging.Tracef("foo %s", "bar")
	logging.Infof("foo %s", "bar")
	logging.Noticef("foo %s", "bar")
	logging.Warnf("foo %s", "bar")
	logging.Errorf("foo %s", "bar")
	logging.Debug("foo bar")
	logging.Trace("foo bar")
	logging.Info("foo bar")
	logging.Notice("foo bar")
	logging.Warn("foo bar")
	logging.Error("foo bar")

	child.End()

	ctx, errSpan := tracer.Start(ctx, "error")

	logging.CtxErrorf(ctx, "error %s", "this is a error")

	logging.Info("no trace context")

	errSpan.End()
}
