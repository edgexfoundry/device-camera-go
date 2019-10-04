.PHONY: build test lint clean prepare update

PKGS := $(shell go list ./... | grep -v /vendor)

GO=CGO_ENABLED=0 GO111MODULE=on go
GOFLAGS=-ldflags

BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter
# APP_PATH must match folder used in Dockerfile
APP_PATH="/usr/local/bin"

MICROSERVICES=./device-camera-go
.PHONY: $(MICROSERVICES)

build: $(MICROSERVICES)
	$(GO) build

test: lint
	$(GO) test ./... -cover

$(GOMETALINTER):
	$(GO) get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

lint: $(GOMETALINTER)
	gometalinter ./... --skip vendor --checkstyle --json --disable gotype --disable gotypex --disable maligned --deadline=200s

clean:
	rm -f $(MICROSERVICES)

prepare:
	$(GO) mod init

update:
	$(GO) mod tidy

docker:
	cp run.sh docker-entrypoint.sh
	# Here we are preserving your local parameters supplied in run.sh to also serve as a default entrypoint for the docker image.
	# These parameters may be overridden by docker run --entrypoint, docker run -e <env var>,
	# or by using docker-compose.yml environment variables
	sed -i '3i #!/bin/sh\n# CAUTION: This file is generated, edits will be overwritten with [make docker] command.\ncd $(APP_PATH)' docker-entrypoint.sh
	chmod +x ./docker-entrypoint.sh
	docker build . --build-arg http_proxy=$(HTTP_PROXY) --build-arg https_proxy=$(HTTPS_PROXY) --tag device-camera-go:develop
