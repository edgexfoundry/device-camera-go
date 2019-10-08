FROM ubuntu:16.04

ARG GIT_COMMIT=unknown
LABEL Name=edgex-device-camera-go Version=0.5.0 git_commit=$GIT_COMMIT

#expose device-camera-go port
ENV APP_PORT=49985

WORKDIR /go/src/github.com/edgexfoundry-holding/device-camera-go

RUN apt-get update && apt-get install -y software-properties-common
RUN add-apt-repository ppa:gophers/archive
RUN apt-get update && apt-get install -y  make git golang-1.11-go
ENV PATH=$PATH:/usr/lib/go-1.11/bin
ENV GOPATH=/go
ENV INSTALL_DIRECTORY=/usr/bin

COPY go.mod .
COPY go.sum .
COPY Makefile .

RUN make update

COPY . .
COPY ./cmd/res/docker/configuration.toml ./cmd/res/configuration.toml
RUN make build

ENTRYPOINT ["./run-docker.sh"]
