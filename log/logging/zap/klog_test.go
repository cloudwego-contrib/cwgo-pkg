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

package zap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/cloudwego/kitex/pkg/klog"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TestLogger test logger work with otelhertz
func TestKLogger(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewKLogger(WithZapOptions(zap.WithFatalHook(zapcore.WriteThenPanic)))
	defer logger.Sync()

	klog.SetLogger(logger)
	klog.SetOutput(buf)
	klog.SetLevel(klog.LevelDebug)

	type logMap map[string]string

	logTestSlice := []logMap{
		{
			"logMessage":       "this is a trace log",
			"formatLogMessage": "this is a trace log: %s",
			"logLevel":         "Trace",
			"zapLogLevel":      "debug",
		},
		{
			"logMessage":       "this is a debug log",
			"formatLogMessage": "this is a debug log: %s",
			"logLevel":         "Debug",
			"zapLogLevel":      "debug",
		},
		{
			"logMessage":       "this is a info log",
			"formatLogMessage": "this is a info log: %s",
			"logLevel":         "Info",
			"zapLogLevel":      "info",
		},
		{
			"logMessage":       "this is a notice log",
			"formatLogMessage": "this is a notice log: %s",
			"logLevel":         "Notice",
			"zapLogLevel":      "warn",
		},
		{
			"logMessage":       "this is a warn log",
			"formatLogMessage": "this is a warn log: %s",
			"logLevel":         "Warn",
			"zapLogLevel":      "warn",
		},
		{
			"logMessage":       "this is a error log",
			"formatLogMessage": "this is a error log: %s",
			"logLevel":         "Error",
			"zapLogLevel":      "error",
		},
		{
			"logMessage":       "this is a fatal log",
			"formatLogMessage": "this is a fatal log: %s",
			"logLevel":         "Fatal",
			"zapLogLevel":      "fatal",
		},
	}

	testHertzLogger := reflect.ValueOf(logger)

	for _, v := range logTestSlice {
		t.Run(v["logLevel"], func(t *testing.T) {
			if v["logLevel"] == "Fatal" {
				defer func() {
					assert.Equal(t, "this is a fatal log", recover())
				}()
			}
			logFunc := testHertzLogger.MethodByName(v["logLevel"])
			logFunc.Call([]reflect.Value{
				reflect.ValueOf(v["logMessage"]),
			})
			assert.Contains(t, buf.String(), v["logMessage"])
			assert.Contains(t, buf.String(), v["zapLogLevel"])

			buf.Reset()

			logfFunc := testHertzLogger.MethodByName(fmt.Sprintf("%sf", v["logLevel"]))
			logfFunc.Call([]reflect.Value{
				reflect.ValueOf(v["formatLogMessage"]),
				reflect.ValueOf(v["logLevel"]),
			})
			assert.Contains(t, buf.String(), fmt.Sprintf(v["formatLogMessage"], v["logLevel"]))
			assert.Contains(t, buf.String(), v["zapLogLevel"])

			buf.Reset()

			ctx := context.Background()
			ctxLogfFunc := testHertzLogger.MethodByName(fmt.Sprintf("Ctx%sf", v["logLevel"]))
			ctxLogfFunc.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(v["formatLogMessage"]),
				reflect.ValueOf(v["logLevel"]),
			})
			assert.Contains(t, buf.String(), fmt.Sprintf(v["formatLogMessage"], v["logLevel"]))
			assert.Contains(t, buf.String(), v["zapLogLevel"])

			buf.Reset()
		})
	}
}

// TestLogLevel test SetLevel
func TestKLogLevel(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewKLogger()
	defer logger.Sync()

	// output to buffer
	logger.SetOutput(buf)

	logger.Debug("this is a debug log")
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.SetLevel(klog.LevelDebug)

	logger.Debugf("this is a debug log %s", "msg")
	assert.True(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.SetLevel(klog.LevelError)
	logger.Infof("this is a debug log %s", "msg")
	assert.False(t, strings.Contains(buf.String(), "this is a info log"))

	logger.Warnf("this is a warn log %s", "msg")
	assert.False(t, strings.Contains(buf.String(), "this is a warn log"))

	logger.Error("this is a error log %s", "msg")
	assert.True(t, strings.Contains(buf.String(), "this is a error log"))
}

func TestKWithCoreEnc(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewKLogger(WithCoreEnc(zapcore.NewConsoleEncoder(humanEncoderConfig())))
	defer logger.Sync()

	// output to buffer
	logger.SetOutput(buf)

	logger.Infof("this is a info log %s", "msg")
	assert.True(t, strings.Contains(buf.String(), "this is a info log"))
}

func TestKWithCoreWs(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewKLogger(WithCoreWs(zapcore.AddSync(buf)))
	defer logger.Sync()

	logger.Infof("this is a info log %s", "msg")
	assert.True(t, strings.Contains(buf.String(), "this is a info log"))
}

func TestKWithCoreLevel(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewKLogger(WithCoreLevel(zap.NewAtomicLevelAt(zapcore.WarnLevel)))
	defer logger.Sync()

	// output to buffer
	logger.SetOutput(buf)

	logger.Infof("this is a info log %s", "msg")
	assert.False(t, strings.Contains(buf.String(), "this is a info log"))

	logger.Warnf("this is a warn log %s", "msg")
	assert.True(t, strings.Contains(buf.String(), "this is a warn log"))
}

// TestCoreOption test zapcore config option
func TestKCoreOption(t *testing.T) {
	buf := new(bytes.Buffer)

	dynamicLevel := zap.NewAtomicLevel()

	dynamicLevel.SetLevel(zap.InfoLevel)

	logger := NewKLogger(
		WithCores([]CoreConfig{
			{
				Enc: zapcore.NewConsoleEncoder(humanEncoderConfig()),
				Ws:  zapcore.AddSync(os.Stdout),
				Lvl: dynamicLevel,
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer("./all/log.log"),
				Lvl: zap.NewAtomicLevelAt(zapcore.DebugLevel),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer("./debug/log.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == zap.DebugLevel
				}),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer("./info/log.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == zap.InfoLevel
				}),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer("./warn/log.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev == zap.WarnLevel
				}),
			},
			{
				Enc: zapcore.NewJSONEncoder(humanEncoderConfig()),
				Ws:  getWriteSyncer("./error/log.log"),
				Lvl: zap.LevelEnablerFunc(func(lev zapcore.Level) bool {
					return lev >= zap.ErrorLevel
				}),
			},
		}...),
	)
	defer logger.Sync()

	logger.SetOutput(buf)

	logger.Debug("this is a debug log")
	// test log level
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.Error("this is a warn log")
	// test log level
	assert.True(t, strings.Contains(buf.String(), "this is a warn log"))
	// test console encoder result
	assert.True(t, strings.Contains(buf.String(), "\tERROR\t"))

	logger.SetLevel(klog.LevelDebug)
	logger.Debug("this is a debug log")
	assert.True(t, strings.Contains(buf.String(), "this is a debug log"))
}

// TestCoreOption test zapcore config option
func TestKZapOption(t *testing.T) {
	buf := new(bytes.Buffer)

	logger := NewKLogger(
		WithZapOptions(zap.AddCaller()),
	)
	defer logger.Sync()

	logger.SetOutput(buf)

	logger.Debug("this is a debug log")
	assert.False(t, strings.Contains(buf.String(), "this is a debug log"))

	logger.Error("this is a warn log")
	// test caller in log result
	assert.True(t, strings.Contains(buf.String(), "caller"))
}

// TestWithExtraKeys test WithExtraKeys option
func TestKWithExtraKeys(t *testing.T) {
	buf := new(bytes.Buffer)

	log := NewKLogger(WithExtraKeys([]ExtraKey{"requestId"}))
	log.SetOutput(buf)

	ctx := context.WithValue(context.Background(), ExtraKey("requestId"), "123")

	log.CtxInfof(ctx, "%s log", "extra")

	var logStructMap map[string]interface{}

	err := json.Unmarshal(buf.Bytes(), &logStructMap)

	assert.Nil(t, err)

	value, ok := logStructMap["requestId"]

	assert.True(t, ok)
	assert.Equal(t, value, "123")
}

func TestKPutExtraKeys(t *testing.T) {
	logger := NewKLogger(WithExtraKeys([]ExtraKey{"abc"}))

	assert.Contains(t, logger.GetExtraKeys(), ExtraKey("abc"))
	assert.NotContains(t, logger.GetExtraKeys(), ExtraKey("def"))

	logger.PutExtraKeys("def")
	assert.Contains(t, logger.GetExtraKeys(), ExtraKey("def"))
}

func TestKExtraKeyAsStr(t *testing.T) {
	buf := new(bytes.Buffer)
	const v = "value"

	logger := NewKLogger(WithExtraKeys([]ExtraKey{"abc"}))

	logger.SetOutput(buf)

	ctx1 := context.TODO()
	ctx1 = context.WithValue(ctx1, "key1", v) //nolint:staticcheck
	logger.CtxErrorf(ctx1, "%s", "error")

	assert.NotContains(t, buf.String(), v)

	buf.Reset()

	strLogger := NewKLogger(WithExtraKeys([]ExtraKey{"abc"}), WithExtraKeyAsStr())

	strLogger.SetOutput(buf)

	ctx2 := context.TODO()
	ctx2 = context.WithValue(ctx2, "key2", v) //nolint:staticcheck

	strLogger.CtxErrorf(ctx2, "key2", v)

	assert.Contains(t, buf.String(), v)

	buf.Reset()
}
