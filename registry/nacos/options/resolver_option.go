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

type ResolverOptions struct {
	Cluster string
	Group   string
}

// ResolverOption Option is nacos registry-hertz option.
type ResolverOption func(o *ResolverOptions)

// WithResolverCluster with cluster option.
func WithResolverCluster(cluster string) ResolverOption {
	return func(o *ResolverOptions) {
		o.Cluster = cluster
	}
}

// WithResolverGroup with group option.
func WithResolverGroup(group string) ResolverOption {
	return func(o *ResolverOptions) {
		o.Group = group
	}
}
