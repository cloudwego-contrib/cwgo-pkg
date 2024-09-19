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

package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider/promprovider"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	registry := prometheus.NewRegistry()

	mux := http.NewServeMux()

	provider := promprovider.NewPromProvider(":9090",
		promprovider.WithRegistry(registry),
		promprovider.WithServeMux(mux),
		promprovider.WithHttpServer(),
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
	measure.Add(context.Background(), semantic.HTTPCounter, 6, labels...)
	measure.Record(context.Background(), semantic.HTTPLatency, float64(time.Second.Microseconds()), labels...)

	promServerResp, err := http.Get("http://localhost:9090/prometheus")
	if err != nil {
		return
	}
	if promServerResp.StatusCode == http.StatusOK {
		fmt.Print("status is 200\n")
	}

	bodyBytes, err := io.ReadAll(promServerResp.Body)
	if err != nil {
		return
	}
	respStr := string(bodyBytes)
	if strings.Contains(respStr, `counter{http_method="/test",path="/cwgo/provider/promProvider",statusCode="200"} 6`) &&
		strings.Contains(respStr, `latency_sum{http_method="/test",path="/cwgo/provider/promProvider",statusCode="200"} 1e+06`) {
		fmt.Print("record and counter work correctly")
	}
}
