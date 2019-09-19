# Camera Device Service

## About
The EdgeX Camera Device Service is developed to control/communicate ONVIF-compliant cameras accessible via http in an EdgeX deployment

## Tested Devices:
The following devices have been tested in a development environment, but ONVIF-compliant cameras
should be compatible.

* Wisenet XND-6080RV
* Bosch NBN-832V-IP
* Hikvision DS-2CD2342WD-I


## Dependencies:

This device service is built using the [onvif4go](https://github.com/faceterteam/onvif4go) library.
It provides a developer-friendly ONVIF client to use with ONVIF-compliant cameras.


## Build Instruction:

1. Check out device-camera-go

2. Build a docker image by using the following command:  
`
docker build . -t device-camera-go
`

3. Alternatively the device service can be run natively with `make build` and the `./run.sh` script.

By default, the configuration and profile files used by the service are available in __'res'__ folder.
