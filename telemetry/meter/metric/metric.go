// Copyright 2023 CloudWeGo Authors.
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

package metric

import (
	"context"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
)

type Measure interface {
	Counter
	Recorder
	RetryRecorder
}
type Counter interface {
	Inc(ctx context.Context, labels []label.CwLabel) error
	Add(ctx context.Context, value int, labels []label.CwLabel) error
}
type Recorder interface {
	Record(ctx context.Context, value float64, labels []label.CwLabel) error
}

type RetryRecorder interface {
	RetryRecord(ctx context.Context, value float64, labels []label.CwLabel) error
}
