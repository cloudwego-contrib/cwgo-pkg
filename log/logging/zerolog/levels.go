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
	"github.com/cloudwego-contrib/cwgo-pkg/log/logging"
	"github.com/rs/zerolog"
)

var (
	zerologLevels = map[logging.Level]zerolog.Level{
		logging.LevelTrace:  zerolog.TraceLevel,
		logging.LevelDebug:  zerolog.DebugLevel,
		logging.LevelInfo:   zerolog.InfoLevel,
		logging.LevelWarn:   zerolog.WarnLevel,
		logging.LevelNotice: zerolog.WarnLevel,
		logging.LevelError:  zerolog.ErrorLevel,
		logging.LevelFatal:  zerolog.FatalLevel,
	}

	logginglevel = map[zerolog.Level]logging.Level{
		zerolog.TraceLevel: logging.LevelTrace,
		zerolog.DebugLevel: logging.LevelDebug,
		zerolog.InfoLevel:  logging.LevelInfo,
		zerolog.WarnLevel:  logging.LevelWarn,
		zerolog.ErrorLevel: logging.LevelError,
		zerolog.FatalLevel: logging.LevelFatal,
	}
)

// matchHlogLevel map hlog.Level to otelzerolog.Level
func matchlogLevel(level logging.Level) zerolog.Level {
	zlvl, found := zerologLevels[level]

	if found {
		return zlvl
	}

	return zerolog.WarnLevel // Default level
}

// matchZerologLevel map otelzerolog.Level to hlog.Level
func matchZerologLevel(level zerolog.Level) logging.Level {
	hlvl, found := logginglevel[level]

	if found {
		return hlvl
	}

	return logging.LevelWarn // Default level
}
