# Camera Device Service
[![Build Status](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-camera-go/job/main/badge/icon)](https://jenkins.edgexfoundry.org/view/EdgeX%20Foundry%20Project/job/edgexfoundry/job/device-camera-go/job/main/) [![Code Coverage](https://codecov.io/gh/edgexfoundry/device-camera-go/branch/main/graph/badge.svg?token=1aTq7fyLNf)](https://codecov.io/gh/edgexfoundry/device-camera-go) [![Go Report Card](https://goreportcard.com/badge/github.com/edgexfoundry/device-camera-go)](https://goreportcard.com/report/github.com/edgexfoundry/device-camera-go) [![GitHub Latest Dev Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-camera-go?include_prereleases&sort=semver&label=latest-dev)](https://github.com/edgexfoundry/device-camera-go/tags) ![GitHub Latest Stable Tag)](https://img.shields.io/github/v/tag/edgexfoundry/device-camera-go?sort=semver&label=latest-stable) [![GitHub License](https://img.shields.io/github/license/edgexfoundry/device-camera-go)](https://choosealicense.com/licenses/apache-2.0/) ![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/edgexfoundry/device-camera-go) [![GitHub Pull Requests](https://img.shields.io/github/issues-pr-raw/edgexfoundry/device-camera-go)](https://github.com/edgexfoundry/device-camera-go/pulls) [![GitHub Contributors](https://img.shields.io/github/contributors/edgexfoundry/device-camera-go)](https://github.com/edgexfoundry/device-camera-go/contributors) [![GitHub Committers](https://img.shields.io/badge/team-committers-green)](https://github.com/orgs/edgexfoundry/teams/device-camera-go-committers/members) [![GitHub Commit Activity](https://img.shields.io/github/commit-activity/m/edgexfoundry/device-camera-go)](https://github.com/edgexfoundry/device-camera-go/commits)


## About
The EdgeX Camera Device Service is developed to control/communicate ONVIF-compliant cameras accessible via http in an EdgeX deployment

## Tested Devices
The following devices have been tested in a development environment with EdgeX. Only those ONVIF-compliant cameras which support the "onvif_snapshot" method are fully compliant with the EdgeX Camera device service:

* Wisenet XND-6080RV
* Bosch NBN-832V-IP
* Hikvision DS-2CD2342WD-I
* SV3C Wireless WiFi Security Camera

The following cameras are non-compliant:

* TP-Link Tapo C110

## Running the device-camera Service as a Snap

The device service is also available as a snap. Install the snap with the following command:

```
$ sudo snap install edgex-device-camera
```

For more details on the Device Camera Snap, including installation, configuration, please refer to [EdgeX Camera Device Service Snap](https://github.com/edgexfoundry/device-camera-go/tree/main/snap)

For more details on Snap, including EdgeX Snap, viewing logs, security services, please check [Getting Started with Snap](https://docs.edgexfoundry.org/2.0/getting-started/Ch-GettingStartedSnapUsers/)

## 

## Dependencies

This device service is built using the [onvif4go](https://github.com/faceterteam/onvif4go) library.
It provides a developer-friendly ONVIF client to use with ONVIF-compliant cameras.


## Build Instructions

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

1. Issue a request to the EdgeX core metadata service to get the Names for all devices.

        GET http://edgex-core-metadata:59881/api/v2/device/all

2. Issue a request to the EdgeX core metadata service to delete a device by ID.

        DELETE http://edgex-core-metadata:59881/api/v2/device/name/{{device_name_here}}

This will remove the device from EdgeX and, as long as it does not remain in the device list
inside this device service's configuration.toml file, will prevent this device service
from attempting to initialize the device at startup. 


#### Example Usage

There are many ways to interact with this device service. This example shows how to emit a snapshot image from your camera at regular intervals by leveraging EdgeX device service auto events. 

1. For this example we will run this in a docker container, so modify the pre-defined devices found [here](./cmd/res/devices/device.toml) to add your IP camera device. Note that at a minimum you must populate the Address value using the correct IP address for your device.

    ```
    # Pre-defined Camera Device(s)
    [[DeviceList]]
    # Unique device name
    Name = "CasualWatcher001"
    # Using default ONVIF camera profile
    Profile = "camera"
    # Human friendly description
    Description = "Camera on east wall facing the loading dock."
    [DeviceList.Protocols]
        [DeviceList.Protocols.HTTP]
        Address = "10.22.34.144"
    # Emit a CBOR-encoded image from this camera, as an EdgeX event, every 30 seconds
    [[DeviceList.AutoEvents]]
        Interval   = "30s"
        OnChange   = false
        SourceName = "OnvifSnapshot"
    ```

2. Build docker image named *edgexfoundry/device-camera-go:0.0.0-dev*.
    ```
    make docker
    ```

3. Generate compose file with device-camera added by running following commands:
   ```
   git clone https://github.com/edgexfoundry/edgex-compose.git
   git checkout <release-branch> # i.e. hanoi
   cd compose-builder/
   make gen no-secty ds-camera
   ```

4. Launch the service using `docker-compose up device-camera`. You should see informational log entries of CBOR events being emitted after 30 seconds, and each 30 seconds thereafter.
    ``` 
    device-camera     | level=INFO msg="Device Service device-camera-go exists"
    device-camera     | level=INFO msg="*Service Start() called, name=device-camera-go, version=1.0.0"
    device-camera     | level=INFO msg="Listening on port: 59985"
    device-camera     | level=INFO msg="Service started in: 65.358268ms"
    device-camera     | level=INFO Content-Type=application/cbor correlation-id=a452e6b5-75b0-46c5-8558-a1c07269bf42 msg="SendEvent: Pushed event to core data"
    device-camera     | level=INFO Content-Type=application/cbor correlation-id=45e53484-6a7e-41e6-9fd5-794f8a002819 msg="SendEvent: Pushed event to core data"
    ```

#### Integrate Into Applications
Refer to [simple-cbor-filter](https://github.com/edgexfoundry/edgex-examples/tree/master/application-services/custom/simple-cbor-filter),
update the `ValueDescriptors` in `res/configuration.toml` and run the example:
```
[ApplicationSettings]
ValueDescriptors = "onvif_snapshot"
```
