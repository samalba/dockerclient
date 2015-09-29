.PHONY: build shell test-unit


BUILD_ID ?= $(shell git rev-parse --short HEAD 2>/dev/null)
DOCKER_IMAGE := dockerclient-dev:$(BUILD_ID)

VOLUMES := \
	-v $(CURDIR):/go/src/github.com/samalba/dockerclient \
	-v $(CURDIR)/dist/bin:/go/bin \
	-v $(CURDIR)/dist/pkg:/go/pkg

all: binary

build:
	docker build -t $(DOCKER_IMAGE) -f Dockerfile.build .

dist:
	mkdir dist/

shell: dist build
	docker run --rm -ti $(VOLUMES) $(DOCKER_IMAGE) bash

test-unit: build
	docker run --rm -ti $(VOLUMES) $(DOCKER_IMAGE) godep go test -v ./...
