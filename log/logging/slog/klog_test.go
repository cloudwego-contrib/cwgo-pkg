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

package slog

import (
	"bufio"
	"bytes"
	"context"
	"log/slog"

	"os"
	"testing"

	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/stretchr/testify/assert"
)

const (
	logFileNamek = "kitex.log"
)

func TestKLogger(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewKLogger()

	klog.SetLogger(logger)
	klog.SetOutput(buf)
	klog.SetLevel(klog.LevelError)

	klog.Info(infoMsg)
	assert.Equal(t, "", buf.String())

	klog.Error(errorMsg)
	// test SetLevel
	assert.Contains(t, buf.String(), errorMsg)

	buf.Reset()
	f, err := os.OpenFile(logFileNamek, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		t.Error(err)
	}

	defer os.Remove(logFileNamek)

	klog.SetOutput(f)

	klog.Info(infoMsg)
	klog.Error(errorMsg)
	_ = f.Sync()

	readF, err := os.OpenFile(logFileNamek, os.O_RDONLY, 0o400)
	if err != nil {
		t.Error(err)
	}
	line, _ := bufio.NewReader(readF).ReadString('\n')

	// test SetOutput
	assert.Contains(t, line, errorMsg)
}

func TestKWithLevel(t *testing.T) {
	buf := new(bytes.Buffer)
	lvl := &slog.LevelVar{}
	lvl.Set(slog.LevelError)
	logger := NewKLogger(WithLevel(lvl))

	klog.SetLogger(logger)
	klog.SetOutput(buf)

	klog.Notice(infoMsg)
	assert.Equal(t, "", buf.String())

	klog.Error(errorMsg)
	assert.Contains(t, buf.String(), errorMsg)

	buf.Reset()
	klog.SetLevel(klog.LevelDebug)
	klog.Debug(debugMsg)

	assert.Contains(t, buf.String(), debugMsg)
}

func TestKWithHandlerOptions(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewKLogger(WithHandlerOptions(&slog.HandlerOptions{Level: slog.LevelError, ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.MessageKey {
			a.Key = "content"
		}
		return a
	}}))

	klog.SetLogger(logger)
	klog.SetOutput(buf)

	klog.Warn(warnMsg)
	assert.Equal(t, "", buf.String())

	klog.SetLevel(klog.LevelInfo)

	klog.Debug(debugMsg)
	assert.Equal(t, "", buf.String())

	klog.Info(infoMsg)
	assert.Contains(t, buf.String(), infoMsg)
	assert.Contains(t, buf.String(), "content")

	buf.Reset()
	klog.SetLevel(klog.LevelTrace)

	testCase := []struct {
		levelName string
		method    func(...any)
		msg       string
	}{
		{
			"Trace",
			klog.Trace,
			traceMsg,
		},
		{
			"Debug",
			klog.Debug,
			debugMsg,
		},
		{
			"Info",
			klog.Info,
			infoMsg,
		},
		{
			"Notice",
			klog.Notice,
			noticeMsg,
		},
		{
			"Warn",
			klog.Warn,
			warnMsg,
		},
		{
			"Error",
			klog.Error,
			errorMsg,
		},
		{
			"Fatal",
			klog.Fatal,
			fatalMsg,
		},
	}

	for _, tc := range testCase {
		tc.method(tc.msg)
		assert.Contains(t, buf.String(), tc.levelName)
		assert.Contains(t, buf.String(), tc.msg)
		buf.Reset()
	}
}

func TestKWithoutLevel(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewKLogger(WithHandlerOptions(&slog.HandlerOptions{AddSource: true}))

	klog.SetLogger(logger)
	klog.SetOutput(buf)

	klog.CtxInfof(context.TODO(), "hello %s", "otelhertz")
	assert.Contains(t, buf.String(), "source")
}

func TestKWithOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	logger := NewKLogger(WithOutput(buf))
	klog.SetLogger(logger)

	klog.CtxErrorf(context.TODO(), errorMsg)
	assert.Contains(t, buf.String(), errorMsg)
}
