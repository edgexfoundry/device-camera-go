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


## Build Instructions:

1. Clone the device-camera-go repo with the following command:

        git clone https://github.com/edgexfoundry/device-camera-go.git

2. Build a docker image by using the following command:  

        docker build . -t device-camera-go

3. Alternatively the device service can be run natively with `make build` and the `./run.sh` script.

By default, the configuration and profile files used by the service are available in the __'cmd/res'__ folder.  Notably,
the configuration.toml file __cmd/res__ should be changed to reflect the correct device profile and
host/IP address for the camera(s).  The configuration-drive.toml file in the same directory should
be updated with the camera's access credentials.

## Notes

#### Parameters for PUT commands

EdgeX Put commands (as-of writing) take a single parameter.  For commands like "create
a user" or "set the device system time", this is problematic.  To workaround this 
limitation, the device service expects the parameters for several of the PUT commands
to be a single string value containing string-escaped JSON.  Some examples:

```$xslt
{"OnvifDateTime": 
    "{
        \"Hour\":21,
        \"Minute\":35,
        \"Second\":6,
        \"Year\":2019,
        \"Month\":10,
        \"Day\":4
    }"
}
```

```$xslt
{"OnvifUser":
    "{
        \"Username\":\"newadmin2\", 
        \"Password\":\"newadmin234\", 
        \"UserLevel\":\"Operator\"
    }"
}
```


#### Removing a Device from EdgeX

During the course of testing or deployment you may end up with EdgeX devices in the system that
you want to remove.  With this device service potentially starting listener goroutines for its
devices, this becomes more relevant.

To remove devices from EdgeX:

1. Issue a request to the EdgeX core metadata service to get the IDs for all devices.

        GET http://edgex-core-metadata:48081/api/v1/device

2. Issue a request to the EdgeX core metadata service to delete a device by ID.

        DELETE http://edgex-core-metadata:48081/api/v1/device/id/{{device_id_here}}

This will remove the device from EdgeX and, as long as it does not remain in the device list
inside this device service's configuration.toml file, will prevent this device service
from attempting to initialize the device at startup. 