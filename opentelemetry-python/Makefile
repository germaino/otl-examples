#SE_PROXY := "http://gateway.schneider.zscaler.net:80"

mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)/..))
current_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))
DOCKER_USER ?= default
APP_INSIGHT_INSTRUMENTATION_KEY :=

DOCKER_PROXY_BUILD_ARG :=
ifneq ($(HTTP_PROXY),)
  DOCKER_PROXY_BUILD_ARG := --build-arg http_proxy=$(HTTP_PROXY)
endif

ifneq ($(HTTPS_PROXY),)
  DOCKER_PROXY_BUILD_ARG += --build-arg https_proxy=$(HTTPS_PROXY)
endif


DOCKER_REGISTRY := cdmorgacrrnd.azurecr.io
DOCKER_IMAGE_TAG := $(shell git rev-parse --short HEAD)
DOCKER_IMAGE := demoapp/demoapp-log

start:
	cat "otel-collector-config.yaml" | sed "s/{{APP_INSIGHT_INSTRUMENTATION_KEY}}/$(APP_INSIGHT_INSTRUMENTATION_KEY)/g" > /tmp/otel-collector-config.yaml
	docker-compose up -d
down:
	docker-compose down
dev:
	docker run -it --rm -v $(mkfile_path):/home/$(DOCKER_USER) -u $(shell id -u):$(shell id -g) --entrypoint bash demoapp/demoapp-log

build:
	docker build $(DOCKER_PROXY_BUILD_ARG) -t $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) $(DOCKER_IMAGE):latest
	docker tag $(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG) $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG)
	docker tag $(DOCKER_IMAGE):latest $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest

push:
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):$(DOCKER_IMAGE_TAG)
	docker push $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest

run:
	docker run --rm $(DOCKER_REGISTRY)/$(DOCKER_IMAGE):latest

deploy_app:
	cat "deployment.yaml" | sed "s/{{DOCKER_TAG}}/$(DOCKER_IMAGE_TAG)/g" | kubectl apply -n demoapp -f -

deploy_otel:
	cat "otel_deployment.yaml" | sed "s/{{APP_INSIGHT_INSTRUMENTATION_KEY}}/$(APP_INSIGHT_INSTRUMENTATION_KEY)/g" | kubectl apply -n otel -f -
.PHONY: build push run deploy dev
