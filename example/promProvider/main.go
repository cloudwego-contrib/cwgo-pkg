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

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/global"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/semantic"

	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/meter/label"
	"github.com/cloudwego-contrib/cwgo-pkg/telemetry/provider/promprovider"
	"github.com/prometheus/client_golang/prometheus"
)

func main() {
	registry := prometheus.NewRegistry()

	provider := promprovider.NewPromProvider(
		promprovider.WithRegistry(registry),
		promprovider.WithHttpServer(), //Activate Kitex monitoring
		promprovider.WithRPCServer(),  //Activate Hertz monitoring
	)

	provider.Serve(":9090", "/metrics-demo")

	labels := []label.CwLabel{
		{Key: "http_method", Value: "/test"},
		{Key: semantic.LabelStatusCode, Value: "200"},
		{Key: "path", Value: "/cwgo/provider/promProvider"},
	}
	measure := global.GetTracerMeasure()

	// Simulate some processing
	measure.Add(context.Background(), semantic.HTTPCounter, 6, labels...)
	measure.Record(context.Background(), semantic.HTTPLatency, float64(time.Second.Microseconds()), labels...)

	promServerResp, err := http.Get("http://localhost:9090/metrics-demo")
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
	if strings.Contains(respStr, `counter{http_method="/test",http_status_code="200",path="/cwgo/provider/promProvider"} 6`) &&
		strings.Contains(respStr, `latency_sum{http_method="/test",http_status_code="200",path="/cwgo/provider/promProvider"} 1e+06`) {
		fmt.Print("record and counter work correctly")
	}
}
