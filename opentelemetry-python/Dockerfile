FROM python:3.7
ARG user=default
ARG group=default
ARG uid=1001
ARG gid=1001

WORKDIR /app

RUN apt update && apt install -y \
  lsof \
  vim

#RUN pip install prometheus_client opencensus opencensus-ext-requests opencensus-ext-jaeger
RUN pip install \
    prometheus_client requests \
    opentelemetry-api \
    opentelemetry-sdk \
    opentelemetry-instrumentation-flask \
    opentelemetry-instrumentation-requests \
    opentelemetry-exporter-jaeger \
    opentelemetry-exporter-prometheus \
    opentelemetry-exporter-otlp

# Create non root user
RUN groupadd -g ${gid} ${group} \
    && useradd -m -u ${uid} -g ${gid} ${user}
RUN chown -R ${user}:${group} /app

# Run as non root user
USER ${user}

COPY --chown=${user}:${user} demoapp_log.py  /app/demoapp.py
#COPY --chown=${user}:${user} titi.py  /app/demoapp.py
ENTRYPOINT  ["python3", "/app/demoapp.py"]
CMD  ["-v", "2"]
