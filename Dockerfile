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
RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY . .
RUN make update 
RUN sed -i '71i // GetDeviceByName returns device if it exists in EdgeX registration cache.\nfunc (s *Service) GetDeviceByName(name string) (models.Device, error) {\n   device, ok := cache.Devices().ForName(name)\n   if !ok {\n      msg := fmt.Sprintf("Device %s cannot be found in cache", name)\n      common.LoggingClient.Info(msg)\n      return models.Device{}, fmt.Errorf(msg)\n   }\n   return device, nil\n}' vendor/github.com/edgexfoundry/device-sdk-go/manageddevices.go

RUN sed -i '186i // Avoid bad deref during error\n	if onvifError.Inner == nil {\n		return onvifError.Message\n	}\n' vendor/github.com/atagirov/goonvif/Device.go

RUN make build

RUN make run

ENTRYPOINT ["./docker-entrypoint.sh"]
