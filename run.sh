#!/bin/sh
cd $GOPATH/src/github.com/edgexfoundry-holding/device-camera-go

# This runs the binary last compiled using 'make build' 
# and provides your desired parameters.
# You may alternately use go run main.go using same parameters.
# This file is used as docker-entrypoint.sh as well.
# NOTE: Environment variables (using dcg_ prefix) override these parameters. 
# docker run -e dcg_ip=14.23.38.1-140
# or export dcg_ip=14.23.38.1-14
# It allows more tuning depending on your deployment; e.g., docker-compose, etc.
./device-camera-go \
 -source onvif:80 \
 -source axis:554 \
 -interval 30 \
 -ip 192.168.0.1-171 \
 -scanduration 15s

