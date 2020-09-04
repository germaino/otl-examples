#! /usr/local/bin/python

import os
import sys
import logging
import time
import argparse
import requests
import random
import flask

from opentelemetry import trace, metrics
from opentelemetry.sdk.resources import Resource


# traces
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchExportSpanProcessor

from opentelemetry.exporter import jaeger
from opentelemetry.exporter.otlp.trace_exporter import OTLPSpanExporter

#metrics
from opentelemetry.sdk.metrics import Counter, MeterProvider
from opentelemetry.sdk.metrics.export import ConsoleMetricsExporter
from opentelemetry.sdk.metrics.export.controller import PushController
from opentelemetry.exporter.prometheus import PrometheusMetricsExporter

# Not supported in 0.12
#from opentelemetry.exporter.otlp.metrics_exporter import OTLPMetricsExporter

# Flask
from opentelemetry.instrumentation.flask import FlaskInstrumentor
from opentelemetry.instrumentation.requests import RequestsInstrumentor

from opentelemetry import propagators
from opentelemetry.sdk.trace.propagation.b3_format import B3Format

#from prometheus_client import start_http_server, Summary, Counter
from prometheus_client import start_http_server
from prometheus_client import make_wsgi_app

from werkzeug.middleware.dispatcher import DispatcherMiddleware


OTL_TRACE_EXPORTER = [ "jaeger", "otl_collector" ]
OTL_METRICS_EXPORTER = [ "prometheus", "otl_collector" ]

def logger_create(name, stream=None):
    #logging.basicConfig(format='%(asctime)s,%(msecs)d %(levelname)-8s [%(filename)s:%(lineno)d] %(message)s', datefmt='%Y-%m-%d:%H:%M:%S')
    logger = logging.getLogger(name)
    loggerhandler = logging.StreamHandler(stream=stream)
    formater = logging.Formatter('[%(asctime)s,%(msecs)d] %(levelname)-8s [%(filename)s:%(lineno)d] - %(message)s','%Y-%m-%d:%H:%M:%S')
    loggerhandler.setFormatter(formater)
    logger.addHandler(loggerhandler)
    logger.setLevel(logging.INFO)
    return logger


def init():

    logger = logger_create(sys.argv[0], stream=sys.stdout)
    parser = argparse.ArgumentParser(description='Log demo app')

    parser.add_argument('-v', '--verbose', default=0,
            dest='verbose', metavar='level', type=int,
            help="Increase verbosity level: 0 = ERROR, 1 = WARNING, 2 = INFO, 3 = DEBUG. (default: ERROR)")

    parser.add_argument('-m', '--metric-exporter', default='prometheus',
            dest='metric_exporter', type=str, choices=OTL_METRICS_EXPORTER,
            help="Exporter to use {}".format(','.join(OTL_METRICS_EXPORTER)))

    parser.add_argument('-t', '--trace-exporter', default='jaeger',
            dest='trace_exporter', type=str, choices=OTL_TRACE_EXPORTER,
            help="Exporter to use {}".format(','.join(OTL_TRACE_EXPORTER)))

    parser.add_argument('--otlp-metric-exporter-url', default='localhost:55680',
            dest='otlp_metric_exporter_url', type=str,
            help="OTLP Exporter url to use (default: %(default)s)")

    parser.add_argument('--otlp-trace-exporter-url', default='localhost:55680',
            dest='otlp_trace_exporter_url', type=str,
            help="OTLP Exporter url to use (default: %(default)s)")

    args = parser.parse_args()

    # Configure logger for requested verbosity.
    if args.verbose == 0:
        logger.setLevel(logging.ERROR)
    elif args.verbose == 1:
        logger.setLevel(logging.WARNING)
    elif args.verbose == 2:
        logger.setLevel(logging.INFO)
    elif args.verbose == 3:
        logger.setLevel(logging.DEBUG)

    # Resource can be required for some backends, e.g. Jaeger
    # If resource wouldn't be set - traces wouldn't appears in Jaeger
    resource = Resource(labels={
        "service.name": "demoapp"
    })

    trace.set_tracer_provider(TracerProvider(resource=resource))

    if 'OTLP_TRACE_EXPORTER_URL' in os.environ:
        trace_exporter_host = os.environ['OTLP_TRACE_EXPORTER_URL'] + ':55680'
    else:
        trace_exporter_host = args.otlp_trace_exporter_url

    if args.trace_exporter == "console":
        span_exporter = ConsoleSpanExporter()
    if args.trace_exporter == "jaeger":
        # create a JaegerSpanExporter
        span_exporter = jaeger.JaegerSpanExporter(
            service_name='demoapp',
            # configure agent
            agent_host_name=trace_exporter_host,
            agent_port=6831,
        )
    elif args.trace_exporter == "otl_collector":
        span_exporter = OTLPSpanExporter(
            endpoint=trace_exporter_host
        )

    # Create a BatchExportSpanProcessor and add the exporter to it
    span_processor = BatchExportSpanProcessor(span_exporter)

    trace.get_tracer_provider().add_span_processor(span_processor)

    propagators.set_global_httptextformat(B3Format())

    # Start Prometheus client
    if args.metric_exporter == "prometheus":
        start_http_server(port=24231)

    metrics.set_meter_provider(MeterProvider())

    batcher_mode = "stateful"
    meter = metrics.get_meter(__name__, batcher_mode == "stateful")

    if 'OTLP_METRIC_EXPORTER_URL' in os.environ:
        metric_exporter_host = os.environ['OTLP_METRIC_EXPORTER_URL']
    else:
        metric_exporter_host = args.otlp_metric_exporter_url

    if args.metric_exporter == "console":
        metric_exporter = ConsoleMetricsExporter("MyAppPrefix")
    elif args.metric_exporter == "prometheus":
        metric_exporter = PrometheusMetricsExporter("MyAppPrefix")
    elif args.metric_exporter == "otl_collector":
        metric_exporter = OTLPMetricsExporter(
            endpoint=metric_exporter_host
        )
    # controller collects metrics created from meter and exports it via the
    # exporter every interval
    controller = PushController(meter, metric_exporter, 5)
    return logger, args, meter



logger, args, meter = init()
app = flask.Flask(__name__)
FlaskInstrumentor().instrument_app(app)
RequestsInstrumentor().instrument()
requests_counter = meter.create_metric(
    name="requests",
    description="number of requests",
    unit="1",
    value_type=int,
    metric_type=Counter,
    enabled=True
)


@app.route("/")
def home():
    tracer = trace.get_tracer(__name__)
    with tracer.start_as_current_span("example-request"):
        requests.get("http://www.example.com")
    requests_counter.add(1, {"route": "home"})
    logger.info("Hit home !")

    return "hello"

if args.metric_exporter == "prometheus":

    # Add prometheus wsgi middleware to route /metrics requests
    app.wsgi_app = DispatcherMiddleware(app.wsgi_app, {
        '/metrics': make_wsgi_app()
    })
app.run(debug=False, port=5000, host='0.0.0.0')
