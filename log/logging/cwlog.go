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

package logging

import (
	"context"
	"fmt"
	"io"
)

type Cwlog struct {
	logger NewLogger
	prefix []CwFeild //全局value
}

func (ll *Cwlog) SetLogger(logger NewLogger) {
	ll.logger = logger
}

func (ll *Cwlog) WithValue(fields ...CwFeild) {
	ll.prefix = append(ll.prefix, fields...)
}

func (ll *Cwlog) ResetValue() {
	ll.prefix = make([]CwFeild, 0)
}

func (ll *Cwlog) SetOutput(w io.Writer) {
	ll.logger.SetOutput(w)
}

func (ll *Cwlog) SetLevel(lv Level) {
	ll.logger.SetLevel(lv)
}

func (ll *Cwlog) Fatal(v ...interface{}) {
	ll.logger.Logw(LevelFatal, fmt.Sprint(v...), ll.prefix...)
}

func (ll *Cwlog) Error(v ...interface{}) {
	ll.logger.Logw(LevelError, fmt.Sprint(v...), ll.prefix...)
}

func (ll *Cwlog) Warn(v ...interface{}) {
	ll.logger.Logw(LevelWarn, fmt.Sprint(v...), ll.prefix...)
}

func (ll *Cwlog) Notice(v ...interface{}) {
	ll.logger.Logw(LevelNotice, fmt.Sprint(v...), ll.prefix...)
}

func (ll *Cwlog) Info(v ...interface{}) {
	ll.logger.Logw(LevelInfo, fmt.Sprint(v...), ll.prefix...)
}

func (ll *Cwlog) Debug(v ...interface{}) {
	ll.logger.Logw(LevelDebug, fmt.Sprint(v...), ll.prefix...)
}

func (ll *Cwlog) Trace(v ...interface{}) {
	ll.logger.Logw(LevelTrace, fmt.Sprint(v...), ll.prefix...)
}

func (ll *Cwlog) Fatalf(format string, v ...interface{}) {
	ll.logger.Logw(LevelFatal, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) Errorf(format string, v ...interface{}) {
	ll.logger.Logw(LevelError, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) Warnf(format string, v ...interface{}) {
	ll.logger.Logw(LevelWarn, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) Noticef(format string, v ...interface{}) {
	ll.logger.Logw(LevelNotice, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) Infof(format string, v ...interface{}) {
	ll.logger.Logw(LevelInfo, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) Debugf(format string, v ...interface{}) {
	ll.logger.Logw(LevelDebug, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) Tracef(format string, v ...interface{}) {
	ll.logger.Logw(LevelTrace, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelFatal, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelError, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelWarn, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelNotice, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelInfo, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelDebug, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelTrace, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *Cwlog) Fatalw(msg string, fields ...CwFeild) {
	kvs := make([]CwFeild, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelFatal, msg, kvs...)
}

func (ll *Cwlog) Errorw(msg string, fields ...CwFeild) {
	kvs := make([]CwFeild, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelError, msg, kvs...)
}

func (ll *Cwlog) Warnw(msg string, fields ...CwFeild) {
	kvs := make([]CwFeild, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelWarn, msg, kvs...)
}

func (ll *Cwlog) Noticew(msg string, fields ...CwFeild) {
	kvs := make([]CwFeild, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelNotice, msg, kvs...)
}

func (ll *Cwlog) Infow(msg string, fields ...CwFeild) {
	kvs := make([]CwFeild, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelInfo, msg, kvs...)
}

func (ll *Cwlog) Debugw(msg string, fields ...CwFeild) {
	kvs := make([]CwFeild, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelDebug, msg, kvs...)
}

func (ll *Cwlog) Tracew(msg string, fields ...CwFeild) {
	kvs := make([]CwFeild, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelTrace, msg, kvs...)
}
