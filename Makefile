.PHONY: build test lint clean prepare update docker

PKGS := $(shell go list ./... | grep -v /vendor)

GO=CGO_ENABLED=0 GO111MODULE=on go
GOFLAGS=-ldflags

BIN_DIR := $(GOPATH)/bin
GOMETALINTER := $(BIN_DIR)/gometalinter

MICROSERVICES=./device-camera-go
.PHONY: $(MICROSERVICES)

build: $(MICROSERVICES)
	go build

test: lint
	go test ./... -cover

$(GOMETALINTER):
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install &> /dev/null

lint: $(GOMETALINTER)
	gometalinter ./... --skip vendor --checkstyle --json --disable gotype --disable gotypex --disable maligned --deadline=200s

clean:
	rm -f $(MICROSERVICES)

prepare:
	go mod init

update:
<<<<<<< HEAD
	dep ensure -update

docker:
	docker build . --build-arg http_proxy=$(HTTP_PROXY) --build-arg https_proxy=$(HTTPS_PROXY) --tag device-camera-go:develop
=======
	go mod tidy
>>>>>>> Upgrade to Edinburgh 1.0
