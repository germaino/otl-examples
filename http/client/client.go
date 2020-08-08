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
        "go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/api/standard"
	"go.opentelemetry.io/otel/api/trace"
        apitrace "go.opentelemetry.io/otel/api/trace"

	"net/http"
	"time"

	"go.opentelemetry.io/otel/api/correlation"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/exporters/otlp"
        "go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
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

        pusher := push.New(
		simple.NewWithExactDistribution(),
		exporter,
		push.WithPeriod(2*time.Second),
	)
	global.SetTraceProvider(tp)
        global.SetMeterProvider(pusher.Provider())
        pusher.Start()
}

func sayHello1(url *string) {

  client := http.DefaultClient
  ctx := correlation.NewContext(context.Background(),
    kv.String("username", "donuts"),
  )

  var body []byte

  tracer := global.Tracer("client/server")
  meter := global.Meter("test-meter")

  fmt.Printf("Sending metrics\n\n\n")
  // labels represent additional key-value descriptors that can be bound to a
  // metric observer or recorder.
  commonLabels := []kv.KeyValue{
      kv.String("labelA", "chocolate"),
      kv.String("labelB", "raspberry"),
      kv.String("labelC", "vanilla"),
  }

  // Recorder metric example
  valuerecorder := metric.Must(meter).
	  NewFloat64Counter(
		"an_important_metric",
		metric.WithDescription("Measures the cumulative epicness of the app"),
	  ).Bind(commonLabels...)
  for i := 0; i < 10; i++ {
    log.Printf("Doing really hard work (%d / 10)\n", i+1)
    valuerecorder.Add(ctx, 1.0)
    time.Sleep(time.Second)
  }
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

func sayHello2() {

  tracer := global.Tracer("ex.com/basic")
  fmt.Printf("Creating child spans\n\n\n")

  tracer.WithSpan(context.Background(), "foo",
    func(ctx context.Context) error {
      tracer.WithSpan(ctx, "bar",
        func(ctx context.Context) error {
          tracer.WithSpan(ctx, "baz",
            func(ctx context.Context) error {
              return nil
            },
          )
          return nil
        },
      )
      return nil
    },
  )
  log.Printf("Done!")
}

func sayHello3() {
  fmt.Printf("Creating span\n\n\n")
  tracer := global.Tracer("toto")
  ctx, span := tracer.Start(context.Background(), "run")
  span.SetAttributes(kv.String("platform", "osx"))
  span.SetAttributes(kv.String("version", "1.2.3"))
  fmt.Printf("Add span event\n\n\n")
  span.AddEvent(ctx, "event in foo", kv.String("name", "foo1"))

  attributes := []kv.KeyValue{
   kv.String("platform", "osx"),
   kv.String("version", "1.2.3"),
  }

  ctx, child := tracer.Start(ctx, "baz", apitrace.WithAttributes(attributes...))
  child.End()
  log.Printf("Done!")
  span.End()
}


func sayHello4(url *string) {

  client := http.DefaultClient
  ctx := correlation.NewContext(context.Background(),
    kv.String("username", "donuts"),
  )

  var body []byte

  tracer := global.Tracer("example/client")
  fmt.Printf("Trace an HTTP request\n\n\n")

  ctx, span := tracer.Start(context.Background(), "say hello4")
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
  //sayHello2()
  //sayHello3()
  sayHello4(url)


}
