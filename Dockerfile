#
# Copyright (c) 2020 IOTech Ltd
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

ARG BASE=golang:1.16-alpine3.12
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

FROM alpine:3.12

COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/cmd /
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/LICENSE /
COPY --from=builder /go/src/github.com/edgexfoundry/device-camera-go/Attribution.txt /

ENTRYPOINT ["/device-camera-go"]
CMD ["--cp=consul://edgex-core-consul:8500", "--registry", "--confdir=/res"]
