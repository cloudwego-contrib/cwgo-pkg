package propagation

import (
	"context"
	"github.com/bytedance/gopkg/cloud/metainfo"
	"github.com/cloudwego-contrib/obs-opentelemetry/tracing/hertz"
	"github.com/cloudwego-contrib/obs-opentelemetry/tracing/kitex"
	"github.com/cloudwego/hertz/pkg/protocol"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

var _ propagation.TextMapCarrier = &metadataProviderHttp{}

type metadataProviderHttp struct {
	metadata map[string]string
	headers  *protocol.RequestHeader
}

// Get a value from metadata by key
func (m *metadataProviderHttp) Get(key string) string {
	return m.headers.Get(key)
}

// Set a value to metadata by k/v
func (m *metadataProviderHttp) Set(key, value string) {
	m.headers.Set(key, value)
}

// Keys Iteratively get all keys of metadata
func (m *metadataProviderHttp) Keys() []string {
	out := make([]string, 0, len(m.metadata))

	m.headers.VisitAll(func(key, value []byte) {
		out = append(out, string(key))
	})

	return out
}
func getMetadataProviderHTTP(headers *protocol.RequestHeader) *metadataProviderHttp {
	return &metadataProviderHttp{headers: headers}
}

// Inject injects span context into the hertz metadata info
func InjectHTTP(ctx context.Context, c *hertz.Config, headers *protocol.RequestHeader) {
	c.GetTextMapPropagator().Inject(ctx, getMetadataProviderHTTP(headers))
}

// Extract returns the baggage and span context
func ExtractHTTP(ctx context.Context, c *hertz.Config, headers *protocol.RequestHeader) (baggage.Baggage, trace.SpanContext) {
	ctx = c.GetTextMapPropagator().Extract(ctx, getMetadataProviderHTTP(headers))
	return baggage.FromContext(ctx), trace.SpanContextFromContext(ctx)
}

var _ propagation.TextMapCarrier = &metadataProviderRPC{}

type metadataProviderRPC struct {
	metadata map[string]string
}

// Get a value from metadata by key
func (m *metadataProviderRPC) Get(key string) string {
	if v, ok := m.metadata[key]; ok {
		return v
	}
	return ""
}

// Set a value to metadata by k/v
func (m *metadataProviderRPC) Set(key, value string) {
	m.metadata[key] = value
}

// Keys Iteratively get all keys of metadata
func (m *metadataProviderRPC) Keys() []string {
	out := make([]string, 0, len(m.metadata))
	for k := range m.metadata {
		out = append(out, k)
	}
	return out
}
func getMetadataProviderRPC(metadata map[string]string) *metadataProviderRPC {
	return &metadataProviderRPC{metadata: metadata}
}

// Inject injects span context into the kitex metadata info
func InjectRPC(ctx context.Context, c *kitex.Config, metadata map[string]string) {
	c.GetTextMapPropagator().Inject(ctx, getMetadataProviderRPC(metadata))
}

// Extract returns the baggage and span context
func ExtractRPC(ctx context.Context, c *kitex.Config, metadata map[string]string) (baggage.Baggage, trace.SpanContext) {
	ctx = c.GetTextMapPropagator().Extract(ctx, getMetadataProviderRPC(CGIVariableToHTTPHeaderMetadata(metadata)))
	return baggage.FromContext(ctx), trace.SpanContextFromContext(ctx)
}

// CGIVariableToHTTPHeaderMetadata converts all CGI variable into HTTP header key.
// For example, `ABC_DEF` will be converted to `abc-def`.
func CGIVariableToHTTPHeaderMetadata(metadata map[string]string) map[string]string {
	res := make(map[string]string, len(metadata))
	for k, v := range metadata {
		res[metainfo.CGIVariableToHTTPHeader(k)] = v
	}
	return res
}

// ExtractFromPropagator get metadata from propagator
func ExtractFromPropagator(ctx context.Context) map[string]string {
	metadata := metainfo.GetAllValues(ctx)
	if metadata == nil {
		metadata = make(map[string]string)
	}
	otel.GetTextMapPropagator().Inject(ctx, getMetadataProviderRPC(metadata))
	return metadata
}
