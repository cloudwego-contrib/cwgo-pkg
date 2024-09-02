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

package promprovider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestPromProvider(t *testing.T) {
	registry := prometheus.NewRegistry()

	mux := http.NewServeMux()

	provider := NewPromProvider(":9090",
		WithRegistry(registry),
		WithServeMux(mux),
		WithCounter(),
		WithRecorder(),
	)
	defer provider.Shutdown(context.Background())
	// assert.NoError(t, err, "Failed to register opsProcessed counter")
	labels := []label.CwLabel{
		{Key: "http_method", Value: "/test"},
		{Key: "statusCode", Value: "200"},
		{Key: "path", Value: "/cwgo/provider/promProvider"},
	}
	measure := provider.Measure
	// 模拟一些处理
	assert.True(t, measure.Add(context.Background(), 6, labels) == nil)
	assert.True(t, measure.Record(context.Background(), float64(time.Second.Microseconds()), labels) == nil)

	promServerResp, err := http.Get("http://localhost:9090/prometheus")
	if err != nil {
		t.Fatal(err)
	}
	assert.True(t, promServerResp.StatusCode == http.StatusOK)

	bodyBytes, err := io.ReadAll(promServerResp.Body)
	assert.True(t, err == nil)
	respStr := string(bodyBytes)
	assert.True(t, strings.Contains(respStr, `counter{http_method="/test",path="/cwgo/provider/promProvider",statusCode="200"} 6`))
	assert.True(t, strings.Contains(respStr, `recorder_sum{http_method="/test",path="/cwgo/provider/promProvider",statusCode="200"} 1e+06`))
}
