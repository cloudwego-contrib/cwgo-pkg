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
	Cluster string
	Group   string
}

// Option is nacos option.
type Option func(o *Options)

// WithCluster with cluster option.
func WithCluster(cluster string) Option {
	return func(o *Options) { o.Cluster = cluster }
}

// WithGroup with group option.
func WithGroup(group string) Option {
	return func(o *Options) { o.Group = group }
}
