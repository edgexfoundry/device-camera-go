FROM golang:1.11.5-alpine3.7 AS builder

ARG GIT_COMMIT=unknown
ENV APP_PORT=49990

LABEL Name=edgex-device-camera-go Version=0.5.0 git_commit=$GIT_COMMIT

#TODO: Assign workdir using relative path such as through $PWD
ENV GO111MODULE=on
WORKDIR ${GOPATH}/src/github.com/edgexfoundry-holding/device-camera-go
LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018, 2019: Intel'

RUN echo http://nl.alpinelinux.org/alpine/v3.7/main > /etc/apk/repositories; \
    echo http://nl.alpinelinux.org/alpine/v3.7/community >> /etc/apk/repositories

RUN apk add --update --no-cache nmap nmap-nselibs nmap-scripts make git bash

COPY go.mod .
COPY provider/go.mod ./provider/go.mod
RUN go mod download

COPY . .

# Consider forcing an update here to handle cases where developer doesn't separate perform step locally
#RUN make update
RUN make build

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019 Intel Corporation'

RUN sed -i '186i // Avoid bad deref during error\n	if onvifError.Inner == nil {\n		return onvifError.Message\n	}\n' vendor/github.com/atagirov/goonvif/Device.go

WORKDIR /

COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/device-camera-go /usr/local/bin/device-camera-go
COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/docker-entrypoint.sh /
COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/res/docker/configuration.toml /res/docker/configuration.toml
COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/res/configuration.toml /res/local/configuration.toml
COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/res/CamDeviceProfile-axis.yaml /res/CamDeviceProfile-axis.yaml
COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/res/CamDeviceProfile.yaml /res/CamDeviceProfile.yaml
COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/res/tags.json /res/tags.json
COPY --from=builder /go/src/github.com/edgexfoundry-holding/device-camera-go/res/cam-credentials.conf /res/cam-credentials.conf

# Copy all for full image useful for interactive session to identify dependencies
#COPY --from=builder . /

# Build ~375MB image getting primary paths
COPY --from=builder /usr ./usr/
COPY --from=builder /bin ./bin/
COPY --from=builder /lib ./lib/

ENV PATH="/usr/local/bin:/usr/bin:/bin:/lib"

# Get minimalist build for <100MB container by identifying/copying explicit dependencies
# Transfer known dependencies for bash, nmap,
#COPY --from=builder /usr/bin/nmap /bin/bash /bin/ls /bin/
#COPY --from=builder /lib/ld-musl-x86_64.so.1 /usr/lib/libreadline.so.7 /lib/ld-musl-x86_64.so.1 /usr/lib/libncursesw.so.6 /lib/ld-musl-x86_64.so.1 /usr/lib/libpcap.so.1 /lib/libssl.so.44 /lib/libcrypto.so.42 /usr/lib/libstdc++.so.6 /usr/lib/libgcc_s.so.1 /lib/ld-musl-x86_64.so.1 /lib/

# Edinburgh docker-compose file is not yet available.
# Use profile parameter to specify nested "local" subfolder beneath confdir
ENTRYPOINT ["./docker-entrypoint.sh"]
