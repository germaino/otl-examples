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

// Example using the OTLP exporter + collector + third-party backends. For
// information about using the exporter, see:
// https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp?tab=doc#example-package-Insecure
package main

import (
	"context"
	"fmt"
        "flag"
	"log"
	"time"

	"math/rand"

	"google.golang.org/grpc"

	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/api/kv"
	"go.opentelemetry.io/otel/api/metric"
	"go.opentelemetry.io/otel/api/standard"
	"go.opentelemetry.io/otel/exporters/otlp"
	"go.opentelemetry.io/otel/sdk/metric/controller/push"
	"go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Initializes an OTLP exporter, and configures the corresponding trace and
// metric providers.
func initProvider(otlAgentAddr *string) (*otlp.Exporter, *push.Controller) {

	fmt.Printf("Opentelemtry server url: %s\n\n\n", *otlAgentAddr)
	// If the OpenTelemetry Collector is running on a local cluster (minikube or
	// microk8s), it should be accessible through the NodePort service at the
	// `localhost:30080` address. Otherwise, replace `localhost` with the
	// address of your cluster. If you run the app inside k8s, then you can
	// probably connect directly to the service through dns
	exp, err := otlp.NewExporter(
		otlp.WithInsecure(),
		otlp.WithAddress(*otlAgentAddr),
		otlp.WithGRPCDialOption(grpc.WithBlock()), // useful for testing
	)
	handleErr(err, "failed to create exporter")

	traceProvider, err := sdktrace.NewProvider(
		sdktrace.WithConfig(sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
		sdktrace.WithResource(resource.New(
			// the service name used to display traces in backends
			kv.Key(standard.ServiceNameKey).String("test-service"),
		)),
		sdktrace.WithSyncer(exp),
	)
	handleErr(err, "failed to create trace provider")

	pusher := push.New(
		simple.NewWithExactDistribution(),
		exp,
		push.WithPeriod(2*time.Second),
	)

	global.SetTraceProvider(traceProvider)
	global.SetMeterProvider(pusher.Provider())
	pusher.Start()

	return exp, pusher
}

func main() {
	log.Printf("Waiting for connection...")

        otlAgentAddr := flag.String("otlagent", "0.0.0.0:55680", "Opentelemetry agent endpoint")
        flag.Parse()

	exp, pusher := initProvider(otlAgentAddr)
	defer func() { handleErr(exp.Stop(), "failed to stop exporter") }()
	defer pusher.Stop() // pushes any last exports to the receiver

	tracer := global.Tracer("test-tracer")
	meter := global.Meter("test-meter")

	// labels represent additional key-value descriptors that can be bound to a
	// metric observer or recorder.
	commonLabels := []kv.KeyValue{
		kv.String("labelA", "chocolate"),
		kv.String("labelB", "raspberry"),
		kv.String("labelC", "vanilla"),
	}

        // Recorder metric example
        counter := metric.Must(meter).NewInt64Counter("api.hit.count")
        //valuerecorder := metric.Must(meter).
	//  NewFloat64ValueRecorder(
	//	"an_important_metric",
		metric.WithDescription("Measures the latency"),
	  ).Bind(commonLabels...)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	for {
		startTime := time.Now()
                ctx, span := tracer.Start(context.Background(), "Foo")
                span.SetAttributes(kv.String("platform", "osx"))
                span.SetAttributes(kv.String("version", "1.2.3"))
                fmt.Printf("Add span event\n\n\n")
                span.AddEvent(ctx, "event in foo", kv.String("name", "foo1"))

		var sleep int64
		switch modulus := time.Now().Unix() % 5; modulus {
		case 0:
			sleep = rng.Int63n(17001)
		case 1:
			sleep = rng.Int63n(8007)
		case 2:
			sleep = rng.Int63n(917)
		case 3:
			sleep = rng.Int63n(87)
		case 4:
			sleep = rng.Int63n(1173)
		}

		time.Sleep(time.Duration(sleep) * time.Millisecond)

		span.End()

                latencyMs := float64(time.Since(startTime)) / 1e6
		nr := int(rng.Int31n(7))
		for i := 0; i < nr; i++ {
			randLineLength := rng.Int63n(999)
			//stats.Record(ctx, mLineLengths.M(randLineLength))
			fmt.Printf("#%d: LineLength: %dBy\n", i, randLineLength)
		}
                valuerecorder.Record(ctx, latencyMs)
                meter.RecordBatch(
		  ctx,
		  commonLabels,
		  valuerecorder.Measurement(2.0),
		  counter.Measurement(12.0),
	        )


		//stats.Record(ctx, mLatencyMs.M(latencyMs))
		fmt.Printf("Latency: %.3fms\n", latencyMs)

	}
	log.Printf("Done!")
}

func handleErr(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}
