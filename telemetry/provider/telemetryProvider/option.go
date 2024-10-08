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
	provider2 "github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider/otelprovider"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider/promprovider"
)

// Option opts for telemetry tracer provider
type Option interface {
	apply(cfg *config)
}

type option func(cfg *config)

func (fn option) apply(cfg *config) {
	fn(cfg)
}

type config struct {
	provider provider2.Provider
}

func WithOtel(opts ...otelprovider.Option) Option {
	return option(func(cfg *config) {
		cfg.provider = otelprovider.NewOpenTelemetryProvider(opts...)
	})
}

func WithProm(opts ...promprovider.Option) Option {
	return option(func(cfg *config) {
		cfg.provider = promprovider.NewPromProvider(opts...)
	})
}

func newConfig(opts []Option) *config {
	cfg := &config{}

	for _, opt := range opts {
		opt.apply(cfg)
	}

	return cfg
}
