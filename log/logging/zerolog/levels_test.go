/*
 * Copyright 2022 CloudWeGo Authors.
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
	"testing"

	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestMatchlogLevel(t *testing.T) {
	assert.Equal(t, zerolog.TraceLevel, matchlogLevel(logging.LevelTrace))
	assert.Equal(t, zerolog.DebugLevel, matchlogLevel(logging.LevelDebug))
	assert.Equal(t, zerolog.InfoLevel, matchlogLevel(logging.LevelInfo))
	assert.Equal(t, zerolog.WarnLevel, matchlogLevel(logging.LevelWarn))
	assert.Equal(t, zerolog.ErrorLevel, matchlogLevel(logging.LevelError))
	assert.Equal(t, zerolog.FatalLevel, matchlogLevel(logging.LevelFatal))
}

func TestMatchZerologLevel(t *testing.T) {
	assert.Equal(t, logging.LevelTrace, matchZerologLevel(zerolog.TraceLevel))
	assert.Equal(t, logging.LevelDebug, matchZerologLevel(zerolog.DebugLevel))
	assert.Equal(t, logging.LevelInfo, matchZerologLevel(zerolog.InfoLevel))
	assert.Equal(t, logging.LevelWarn, matchZerologLevel(zerolog.WarnLevel))
	assert.Equal(t, logging.LevelError, matchZerologLevel(zerolog.ErrorLevel))
	assert.Equal(t, logging.LevelFatal, matchZerologLevel(zerolog.FatalLevel))
}
