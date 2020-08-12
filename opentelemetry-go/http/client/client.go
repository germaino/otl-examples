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
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/standard"
	"go.opentelemetry.io/otel/api/trace"

	"net/http"
	"time"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
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
	    sdktrace.WithResource(resource.New(standard.ServiceNameKey.String("ClientExample"))),
	    sdktrace.WithBatcher(exporter))
	handleErr(err, "failed to create trace provider")

	global.SetTraceProvider(tp)
}

func sayHello1(url *string) {

  client := http.DefaultClient
  ctx := correlation.NewContext(context.Background(),
    kv.String("username", "donuts"),
  )

  var body []byte

  tracer := global.Tracer("client/server")

  fmt.Printf("Trace an HTTP request\n\n\n")

  err := tracer.WithSpan(ctx, "say hello1",
    func(ctx context.Context) error {
      req, _ := http.NewRequest("GET", *url, nil)

      ctx, req = httptrace.W3C(ctx, req)
      httptrace.Inject(ctx, req)

      fmt.Printf("Sending request...\n")
      res, err := client.Do(req)

      if err != nil {
        panic(err)
      }

      body, err = ioutil.ReadAll(res.Body)
      _ = res.Body.Close()

      return err
    },
    trace.WithAttributes(standard.PeerServiceKey.String("ExampleService")))

    if err != nil {
      panic(err)
    }

    fmt.Printf("Response Received: %s\n\n\n", body)
    fmt.Printf("Waiting for few seconds to export spans ...\n\n")
    time.Sleep(10 * time.Second)
    fmt.Printf("Inspect traces on stdout\n")
    log.Printf("Done!")
}


func sayHello2(url *string) {

  client := http.DefaultClient
  ctx := correlation.NewContext(context.Background(),
    kv.String("username", "donuts"),
  )

  var body []byte

  tracer := global.Tracer("example/client")
  fmt.Printf("Trace an HTTP request\n\n\n")

  ctx, span := tracer.Start(context.Background(), "say hello2")
  req, _ := http.NewRequest("GET", *url, nil)
  ctx, req = httptrace.W3C(ctx, req)
  httptrace.Inject(ctx, req)
  fmt.Printf("Sending request...\n")
  res, err := client.Do(req)

  if err != nil {
    panic(err)
  }

  body, err = ioutil.ReadAll(res.Body)
  _ = res.Body.Close()

  span.End()

  fmt.Printf("Response Received: %s\n\n\n", body)
  fmt.Printf("Waiting for few seconds to export spans ...\n\n")
  time.Sleep(10 * time.Second)
  fmt.Printf("Inspect traces on stdout\n")
  log.Printf("Done!")
}



func main() {
  url := flag.String("server", "http://localhost:7777/hello", "server url")
  otlAgentAddr := flag.String("otlagent", "0.0.0.0:55680", "Opentelemetry agent endpoint")
  flag.Parse()
  log.Printf("Waiting for connection...")
  initProvider(otlAgentAddr)

  sayHello1(url)
  sayHello2(url)


}
