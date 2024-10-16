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

package zerolog

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKFrom(t *testing.T) {
	b := &bytes.Buffer{}

	zl := zerolog.New(b).With().Str("key", "test").Logger()
	l := FromK(zl)

	l.Info("foo")

	assert.Equal(
		t,
		`{"level":"info","key":"test","message":"foo"}
`,
		b.String(),
	)
}

func TestKGetLogger_notSet(t *testing.T) {
	_, err := GetKLogger()

	assert.Error(t, err)
	assert.Equal(t, "klog.DefaultLogger is not a zerolog logger", err.Error())
}

func TestKGetLogger(t *testing.T) {
	klog.SetLogger(NewK())
	logger, err := GetKLogger()

	assert.NoError(t, err)
	assert.IsType(t, KLogger{}, logger)
}

func TestKWithContext(t *testing.T) {
	ctx := context.Background()
	l := NewK()
	c := l.WithContext(ctx)

	assert.NotNil(t, c)
	assert.IsType(t, zerolog.Ctx(c), &zerolog.Logger{})
}

func TestKLoggerWithField(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)
	l.WithField("service", "logging")

	l.Info("foobar")

	type Log struct {
		Level   string `json:"level"`
		Service string `json:"service"`
		Message string `json:"message"`
	}

	log := &Log{}

	err := json.Unmarshal(b.Bytes(), log)

	println(b.String())
	assert.NoError(t, err)
	assert.Equal(t, "logging", log.Service)
}

func TestKUnwrap(t *testing.T) {
	l := NewK()

	logger := l.Unwrap()

	assert.NotNil(t, logger)
	assert.IsType(t, zerolog.Logger{}, logger)
}

func TestKLog(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)

	l.Trace("foo")
	assert.Equal(
		t,
		`{"level":"debug","message":"foo"}
`,
		b.String(),
	)

	b.Reset()
	l.Debug("foo")
	assert.Equal(
		t,
		`{"level":"debug","message":"foo"}
`,
		b.String(),
	)

	b.Reset()
	l.Info("foo")
	assert.Equal(
		t,
		`{"level":"info","message":"foo"}
`,
		b.String(),
	)

	b.Reset()
	l.Notice("foo")
	assert.Equal(
		t,
		`{"level":"warn","message":"foo"}
`,
		b.String(),
	)

	b.Reset()
	l.Warn("foo")
	assert.Equal(
		t,
		`{"level":"warn","message":"foo"}
`,
		b.String(),
	)

	b.Reset()
	l.Error("foo")
	assert.Equal(
		t,
		`{"level":"error","message":"foo"}
`,
		b.String(),
	)
}

func TestKLogf(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)

	l.Tracef("foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"debug","message":"foobar"}
`,
		b.String(),
	)

	b.Reset()
	l.Debugf("foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"debug","message":"foobar"}
`,
		b.String(),
	)

	b.Reset()
	l.Infof("foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"info","message":"foobar"}
`,
		b.String(),
	)

	b.Reset()
	l.Noticef("foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"warn","message":"foobar"}
`,
		b.String(),
	)

	b.Reset()
	l.Warnf("foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"warn","message":"foobar"}
`,
		b.String(),
	)

	b.Reset()
	l.Errorf("foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"error","message":"foobar"}
`,
		b.String(),
	)
}

func TestKCtxTracef(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)
	ctx := l.log.WithContext(context.Background())

	l.CtxTracef(ctx, "foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"debug","message":"foobar"}
`,
		b.String(),
	)
	assert.NotNil(t, log.Ctx(ctx))
}

func TestKCtxDebugf(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)
	ctx := l.log.WithContext(context.Background())

	l.CtxDebugf(ctx, "foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"debug","message":"foobar"}
`,
		b.String(),
	)
	assert.NotNil(t, log.Ctx(ctx))
}

func TestKCtxInfof(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)
	ctx := l.log.WithContext(context.Background())

	l.CtxInfof(ctx, "foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"info","message":"foobar"}
`,
		b.String(),
	)
	assert.NotNil(t, log.Ctx(ctx))
}

func TestKCtxNoticef(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)
	ctx := l.log.WithContext(context.Background())

	l.CtxNoticef(ctx, "foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"warn","message":"foobar"}
`,
		b.String(),
	)
	assert.NotNil(t, log.Ctx(ctx))
}

func TestKCtxWarnf(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)
	ctx := l.log.WithContext(context.Background())

	l.CtxWarnf(ctx, "foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"warn","message":"foobar"}
`,
		b.String(),
	)
	assert.NotNil(t, log.Ctx(ctx))
}

func TestKCtxErrorf(t *testing.T) {
	b := &bytes.Buffer{}
	l := NewK()
	l.SetOutput(b)
	ctx := l.log.WithContext(context.Background())

	l.CtxErrorf(ctx, "foo%s", "bar")
	assert.Equal(
		t,
		`{"level":"error","message":"foobar"}
`,
		b.String(),
	)
	assert.NotNil(t, log.Ctx(ctx))
}

func TestKSetLevel(t *testing.T) {
	l := NewK()

	l.SetLevel(klog.LevelDebug)
	assert.Equal(t, l.log.GetLevel(), zerolog.DebugLevel)

	l.SetLevel(klog.LevelDebug)
	assert.Equal(t, l.log.GetLevel(), zerolog.DebugLevel)

	l.SetLevel(klog.LevelError)
	assert.Equal(t, l.log.GetLevel(), zerolog.ErrorLevel)
}
