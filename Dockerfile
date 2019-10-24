ARG BASE=golang:1.12-alpine
FROM ${BASE} AS builder

LABEL Name=edgex-device-camera-go

#expose device-camera-go port
ENV APP_PORT=49985

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018, 2019: Intel'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache make git

WORKDIR /go/src/github.com/edgexfoundry/device-camera-go

COPY go.mod .
COPY Makefile .

RUN make update

COPY . .
RUN make build

FROM alpine

COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd/res/docker/configuration.toml /cmd/res/configuration.toml
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd/res/camera.yaml /cmd/res/camera.yaml
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd/res/camera-axis.yaml /cmd/res/camera-axis.yaml
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd/res/camera-bosch.yaml /cmd/res/camera-bosch.yaml
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd/res/configuration-driver.toml /cmd/res/configuration-driver.toml
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd/device-camera-go /cmd/device-camera-go
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/run-docker.sh /run-docker.sh

ENTRYPOINT ["/run-docker.sh"]
