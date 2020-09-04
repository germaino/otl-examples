# Overview
Capture metrics using Opentelemtry collector from a metric load generator in opencensus format
Visualize metrics in prometheus
Internal collector metrics can be displayed without prometheus
Original code: from [Github](https://github.com/open-telemetry/opentelemetry-collector/tree/master/examples
)

![use case](./docs/otel-collector-std.png)

# Build
make build

# Run with docker-compose
make start

# Visualize metrics with prometheus
## Text output endpoint
curl http://localhost:8888/metrics
## Inside prometheus
firefox http://localhost:9090

# Stop
make down
