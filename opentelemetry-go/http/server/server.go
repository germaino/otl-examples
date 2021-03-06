// Copyright The OpenTelemetry Authors
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

package main

import (
	"io"
	"log"
	"fmt"
	"flag"
	"net/http"

	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/standard"
	"go.opentelemetry.io/otel/api/trace"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/instrumentation/httptrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func initProvider(otlAgentAddr *string) {


	fmt.Printf("Opentelemtry server url: %s\n\n\n", *otlAgentAddr)
	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` address. Otherwise, replace `localhost` with the
	// address of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	exporter, err := otlp.NewExporter(
		otlp.WithInsecure(),
                otlp.WithAddress(*otlAgentAddr),
		otlp.WithGRPCDialOption(grpc.WithBlock()), // useful for testing
	)
	handleErr(err, "failed to create exporter")

	// For the demonstration, use sdktrace.AlwaysSample sampler to sample all traces.
	// In a production application, use sdktrace.ProbabilitySampler with a desired probability.
	tp, err := sdktrace.NewProvider(
            sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	    sdktrace.WithSyncer(exporter),
	    sdktrace.WithResource(resource.New(standard.ServiceNameKey.String("ServerExample"))))

        handleErr(err, "failed to create trace provider")
	global.SetTraceProvider(tp)
}

func helloHandler(w http.ResponseWriter, req *http.Request) {
      tracer := global.TraceProvider().Tracer("example/server")

       // Extracts the conventional HTTP span attributes,
       // distributed context tags, and a span context for
       // tracing this request.
       attrs, entries, spanCtx := httptrace.Extract(req.Context(), req)
       ctx := req.Context()
       if spanCtx.IsValid() {
             ctx = trace.ContextWithRemoteSpanContext(ctx, spanCtx)
       }

       // Apply the correlation context tags to the request
       // context.
       req = req.WithContext(correlation.ContextWithMap(ctx, correlation.NewMap(correlation.MapUpdate{
             MultiKV: entries,
       })))

       // Start the server-side span, passing the remote
       // child span context explicitly.
       _, span := tracer.Start(
             req.Context(),
             "hello",
             trace.WithAttributes(attrs...),
       )
       defer span.End()

       // Add event
       span.AddEvent(ctx, "handling this...")

       io.WriteString(w, "Hello, world!\n")
}

func main() {
      otlAgentAddr := flag.String("otlagent", "0.0.0.0:55680", "Opentelemetry agent endpoint")
      flag.Parse()
      initProvider(otlAgentAddr)

      http.HandleFunc("/hello", helloHandler)
      err := http.ListenAndServe(":7777", nil)
      if err != nil {
            panic(err)
      }
}


