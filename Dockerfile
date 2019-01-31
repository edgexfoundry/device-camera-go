FROM golang:1.11.2-alpine3.7 AS builder

ARG GIT_COMMIT=unknown
LABEL Name=edgex-device-camera-go Version=0.5.0 git_commit=$GIT_COMMIT

#expose device-camera-go port
ENV APP_PORT=49990

WORKDIR ~/go/src/github.com/edgexfoundry-holding/device-camera-go

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.
RUN echo http://nl.alpinelinux.org/alpine/v3.7/main > /etc/apk/repositories; \
    echo http://nl.alpinelinux.org/alpine/v3.7/community >> /etc/apk/repositories

RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
RUN apk add --update --no-cache nmap nmap-nselibs nmap-scripts make

COPY . .
RUN make ./

FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2018: TBD'

EXPOSE $APP_PORT

COPY --from=builder . /

ENTRYPOINT ["/device-camera-go","-registry","-source onvif","-source axis"]

