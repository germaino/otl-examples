# Build
make build

# Run
make start

# Visualize metrics with prometheus
firefox http://localhost:9090

# Visualize traces
firefox http://localhost:9411/zipkin
firefox http://localhost:16686

# Stop
make down
