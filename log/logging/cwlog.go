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

type CwLog struct {
	logger NewLogger
	prefix []CwField //global value
}

func (ll *CwLog) SetLogger(logger NewLogger) {
	ll.logger = logger
}

func (ll *CwLog) WithValue(fields ...CwField) {
	ll.prefix = append(ll.prefix, fields...)
}

func (ll *CwLog) ResetValue() {
	ll.prefix = make([]CwField, 0)
}

func (ll *CwLog) SetOutput(w io.Writer) {
	ll.logger.SetOutput(w)
}

func (ll *CwLog) SetLevel(lv Level) {
	ll.logger.SetLevel(lv)
}

func (ll *CwLog) Fatal(v ...interface{}) {
	ll.logger.Logw(LevelFatal, fmt.Sprint(v...), ll.prefix...)
}

func (ll *CwLog) Error(v ...interface{}) {
	ll.logger.Logw(LevelError, fmt.Sprint(v...), ll.prefix...)
}

func (ll *CwLog) Warn(v ...interface{}) {
	ll.logger.Logw(LevelWarn, fmt.Sprint(v...), ll.prefix...)
}

func (ll *CwLog) Notice(v ...interface{}) {
	ll.logger.Logw(LevelNotice, fmt.Sprint(v...), ll.prefix...)
}

func (ll *CwLog) Info(v ...interface{}) {
	ll.logger.Logw(LevelInfo, fmt.Sprint(v...), ll.prefix...)
}

func (ll *CwLog) Debug(v ...interface{}) {
	ll.logger.Logw(LevelDebug, fmt.Sprint(v...), ll.prefix...)
}

func (ll *CwLog) Trace(v ...interface{}) {
	ll.logger.Logw(LevelTrace, fmt.Sprint(v...), ll.prefix...)
}

func (ll *CwLog) Fatalf(format string, v ...interface{}) {
	ll.logger.Logw(LevelFatal, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) Errorf(format string, v ...interface{}) {
	ll.logger.Logw(LevelError, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) Warnf(format string, v ...interface{}) {
	ll.logger.Logw(LevelWarn, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) Noticef(format string, v ...interface{}) {
	ll.logger.Logw(LevelNotice, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) Infof(format string, v ...interface{}) {
	ll.logger.Logw(LevelInfo, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) Debugf(format string, v ...interface{}) {
	ll.logger.Logw(LevelDebug, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) Tracef(format string, v ...interface{}) {
	ll.logger.Logw(LevelTrace, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) CtxFatalf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelFatal, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) CtxErrorf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelError, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) CtxWarnf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelWarn, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) CtxNoticef(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelNotice, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) CtxInfof(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelInfo, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) CtxDebugf(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelDebug, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) CtxTracef(ctx context.Context, format string, v ...interface{}) {
	ll.logger.CtxLog(LevelTrace, ctx, fmt.Sprintf(format, v...), ll.prefix...)
}

func (ll *CwLog) Fatalw(msg string, fields ...CwField) {
	kvs := make([]CwField, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelFatal, msg, kvs...)
}

func (ll *CwLog) Errorw(msg string, fields ...CwField) {
	kvs := make([]CwField, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelError, msg, kvs...)
}

func (ll *CwLog) Warnw(msg string, fields ...CwField) {
	kvs := make([]CwField, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelWarn, msg, kvs...)
}

func (ll *CwLog) Noticew(msg string, fields ...CwField) {
	kvs := make([]CwField, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelNotice, msg, kvs...)
}

func (ll *CwLog) Infow(msg string, fields ...CwField) {
	kvs := make([]CwField, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelInfo, msg, kvs...)
}

func (ll *CwLog) Debugw(msg string, fields ...CwField) {
	kvs := make([]CwField, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelDebug, msg, kvs...)
}

func (ll *CwLog) Tracew(msg string, fields ...CwField) {
	kvs := make([]CwField, 0, len(ll.prefix)+len(fields))

	kvs = append(kvs, ll.prefix...)
	kvs = append(kvs, fields...)
	ll.logger.Logw(LevelTrace, msg, kvs...)
}
