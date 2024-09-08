// Copyright 2022 CloudWeGo Authors.
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

package semantic

import (
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
)

type ServiceType string

const (
	Hertz = ServiceType("hertz")
	Kitex = ServiceType("Kitex")
)

const (
	Counter = "counter"
	Latency = "latency"
	Retry   = "retry"
)

// RPC measure Labels
const (
	LabelRPCMethodKey = "rpc_method"
	LabelRPCCalleeKey = "rpc_service"
	LabelRPCCallerKey = "caller_rpc_service"
	LabelKeyRetry     = "retry"
	LabelKeyStatus    = "status"
)

// HTTP measure Labels
const (
	LabelHttpMethodKey = "http_method"
	LabelStatusCode    = "statusCode"
	LabelPath          = "path"
)

// common Labels
const (
	LabelMethod       = "method"
	UnknownLabelValue = "unknown"
	StatusSucceed     = "succeed"
	StatusError       = "error"
)

// otel keys
const (
	// RequestProtocolKey protocol of the request.
	//
	// Type: string
	// Required: Always
	// Examples:
	// http: 'http'
	// rpc: 'grpc', 'java_rmi', 'wcf', 'otelkitex'
	// db: mysql, postgresql
	// mq: 'rabbitmq', 'activemq', 'AmazonSQS'
	RequestProtocolKey = attribute.Key("request.protocol")

	// RPCSystemKitexRecvSize recv_size
	RPCSystemKitexRecvSize = attribute.Key("otelkitex.recv_size")
	// RPCSystemKitexSendSize send_size
	RPCSystemKitexSendSize = attribute.Key("otelkitex.send_size")

	// PeerServiceNamespaceKey peer.service.namespace
	PeerServiceNamespaceKey = attribute.Key("peer.service.namespace")
	// PeerDeploymentEnvironmentKey peer.deployment.environment
	PeerDeploymentEnvironmentKey = attribute.Key("peer.deployment.environment")
)

const (
	// SourceOperationKey source operation
	//
	// Type: string
	// Required: Optional
	// Examples: '/operation1'
	SourceOperationKey = attribute.Key("source_operation")
)

const (
	StatusKey = attribute.Key("status.code")
)

// RPC Server meter
// ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/semantic_conventions/rpc.md#rpc-server
const (
	SeverThroughput       = "rpc.server.throughput"
	ServerDuration        = "rpc.server.duration"          // measures duration of inbound RPC
	ServerRequestSize     = "rpc.server.request.size"      // measures size of RPC request messages (uncompressed)
	ServerResponseSize    = "rpc.server.response.size"     // measures size of RPC response messages (uncompressed)
	ServerRequestsPerRPC  = "rpc.server.requests_per_rpc"  // measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs
	ServerResponsesPerRPC = "rpc.server.responses_per_rpc" // measures the number of messages sent per RPC. Should be 1 for all non-streaming RPCs
	ServerRetry           = "rpc.server.retry"
)

// RPC Client meter
// ref to https://github.com/open-telemetry/opentelemetry-specification/blob/main/specification/metrics/semantic_conventions/rpc.md#rpc-client
const (
	ClientThroughput      = "rpc.client.throughput"
	ClientDuration        = "rpc.client.duration"          // measures duration of outbound RPC
	ClientRequestSize     = "rpc.client.request.size"      // measures size of RPC request messages (uncompressed)
	ClientResponseSize    = "rpc.client.response.size"     // measures size of RPC response messages (uncompressed)
	ClientRequestsPerRPC  = "rpc.client.requests_per_rpc"  // measures the number of messages received per RPC. Should be 1 for all non-streaming RPCs
	ClientResponsesPerRPC = "rpc.client.responses_per_rpc" // measures the number of messages sent per RPC. Should be 1 for all non-streaming RPCs
	ClientRetry           = "rpc.client.retry"
)

// Server HTTP meter
const (
	ServerRequestCount = "http.server.request_count" // measures the incoming request count total
	ServerLatency      = "http.server.duration"      // measures th incoming end to end duration
)

// Client HTTP meter.
const (
	ClientRequestCount = "http.client.request_count" // measures the client request count total
	ClientLatency      = "http.client.duration"      // measures the duration outbound HTTP requests
)

// RPCSystemKitex Semantic convention for otelkitex as the remoting system.
var RPCSystemKitex = semconv.RPCSystemKey.String("kitex")
