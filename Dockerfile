ARG BASE=golang:1.13-alpine
FROM ${BASE} AS builder

ARG MAKE="make build"
ARG ALPINE_PKG_BASE="make git"
ARG ALPINE_PKG_EXTRA=""

LABEL Name=edgex-device-camera-go

#expose device-camera-go port
ENV APP_PORT=49985

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018-2020: Intel'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}

WORKDIR $GOPATH/src/github.com/edgexfoundry/device-camera-go

COPY go.mod .
COPY Makefile .

RUN make update

COPY . .
RUN ${MAKE}

FROM alpine

COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd /
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/LICENSE /
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/Attribution.txt /

ENTRYPOINT ["/device-camera-go","--cp=consul://edgex-core-consul:8500","--registry","--confdir=/res"]
