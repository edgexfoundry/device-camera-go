#!/bin/sh

cd cmd || exit
./device-camera-go -r consul://edgex-core-consul:8500 -c res/
