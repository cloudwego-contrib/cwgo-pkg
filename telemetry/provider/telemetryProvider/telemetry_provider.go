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

package telemetryProvider

import (
	"context"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider"
)

var _ provider.Provider = &TelemetryProvider{}

type TelemetryProvider struct {
	provider provider.Provider
}

func (t TelemetryProvider) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}

func NewTelemetryProvider(opts ...Option) provider.Provider {
	cfg := newConfig(opts)

	return &TelemetryProvider{cfg.provider}
}
