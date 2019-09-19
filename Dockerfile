FROM ubuntu:16.04

ARG GIT_COMMIT=unknown
LABEL Name=edgex-device-camera-go Version=0.5.0 git_commit=$GIT_COMMIT

#expose device-camera-go port
ENV APP_PORT=49990

#TODO: Take $PWD as input?
WORKDIR /go/src/github.com/edgexfoundry-holding/device-camera-go

RUN apt update && apt install -y software-properties-common
RUN add-apt-repository ppa:gophers/archive
RUN apt update && apt install -y nmap make git curl golang-1.11-go
ENV PATH=$PATH:/usr/lib/go-1.11/bin
ENV GOPATH=/go
ENV INSTALL_DIRECTORY=/usr/bin

COPY go.mod .
COPY go.sum .
COPY Makefile .

RUN make update

COPY . .

RUN make build

ENTRYPOINT ["./run.sh"]