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

package options

type Options struct {
	AppId             string
	VersionRule       string
	HostName          string
	HeartbeatInterval int32
}

// Option is ServiceComb option.
type Option func(o *Options)

// WithAppId with app id option
func WithAppId(appId string) Option {
	return func(o *Options) {
		o.AppId = appId
	}
}

// WithVersionRule with version rule option
func WithVersionRule(versionRule string) Option {
	return func(o *Options) {
		o.VersionRule = versionRule
	}
}

// WithHostName with host name option
func WithHostName(hostName string) Option {
	return func(o *Options) {
		o.HostName = hostName
	}
}

func WithHeartbeatInterval(second int32) Option {
	return func(o *Options) {
		o.HeartbeatInterval = second
	}
}
