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

package global

import (
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/metric"
	"sync"
)

// Global variable, used to store the current TracerProvider
var (
	tracerMeasure metric.Measure = metric.NewMeasure()
	lock          sync.Mutex
)

// SetTracerMeasure used to set TracerProvider
func SetTracerMeasure(measure metric.Measure) {
	lock.Lock()
	defer lock.Unlock()
	tracerMeasure = measure
}

// GetTracerMeasure used to get TracerProvider
func GetTracerMeasure() metric.Measure {
	lock.Lock()
	defer lock.Unlock()
	return tracerMeasure
}
