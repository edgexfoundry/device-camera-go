# Camera Device Service

## About
The EdgeX Camera Device Service is developed to control/communicate ONVIF-compliant cameras accessible via http in an EdgeX deployment

## Tested Devices
The following devices have been tested in a development environment, but ONVIF-compliant cameras
should be compatible.

* Wisenet XND-6080RV
* Bosch NBN-832V-IP
* Hikvision DS-2CD2342WD-I


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

1. Issue a request to the EdgeX core metadata service to get the IDs for all devices.

        GET http://edgex-core-metadata:48081/api/v1/device

2. Issue a request to the EdgeX core metadata service to delete a device by ID.

        DELETE http://edgex-core-metadata:48081/api/v1/device/id/{{device_id_here}}

This will remove the device from EdgeX and, as long as it does not remain in the device list
inside this device service's configuration.toml file, will prevent this device service
from attempting to initialize the device at startup. 


#### Example Usage

There are many ways to interact with this device service. This example shows how to emit a snapshot image from your camera at regular intervals by leveraging EdgeX device service auto events. 

1. For this example we will run this in a docker container, so modify the device service configuration profile for docker found [here](./cmd/res/docker/configuration.toml) to add your IP camera device. Note that at a minimum you must populate the Address value using the correct IP address for your device.

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
        Frequency = "30s"
        OnChange = false
        Resource = "onvif_snapshot"
    ```

2. Build docker image named *device-camera-go*.
    ```
    make docker
    ```

3. Save the EdgeX docker-compose template found [here](https://github.com/edgexfoundry/developer-scripts/blob/master/releases/edinburgh/compose-files/) as docker-compose.yml. Then update your docker-compose.yml file to add this device service to the stack.
    ```
    #################################################################
    # Device Services
    #################################################################
    docker-device-camera-go:
        image: device-camera-go:latest
        ports:
        - "49985:49985"
        container_name: edgex-device-camera-go
        hostname: edgex-device-camera-go
        networks:
        edgex-network:
            aliases:
            - edgex-device-camera-go
        volumes:
        - db-data:/data/db
        - log-data:/edgex/logs
        - consul-config:/consul/config
        - consul-data:/consul/data
        depends_on:
        - data
        - command
        - metadata
    ```

4. Launch the service using `docker-compose up docker-device-camera-go`. You should see informational log entries of CBOR events being emitted after 30 seconds, and each 30 seconds thereafter.
    ``` 
    edgex-device-camera-go     | level=INFO msg="Device Service device-camera-go exists"
    edgex-device-camera-go     | level=INFO msg="*Service Start() called, name=device-camera-go, version=1.0.0"
    edgex-device-camera-go     | level=INFO msg="Listening on port: 49985"
    edgex-device-camera-go     | level=INFO msg="Service started in: 65.358268ms"
    edgex-device-camera-go     | level=INFO Content-Type=application/cbor correlation-id=a452e6b5-75b0-46c5-8558-a1c07269bf42 msg="SendEvent: Pushed event to core data"
    edgex-device-camera-go     | level=INFO Content-Type=application/cbor correlation-id=45e53484-6a7e-41e6-9fd5-794f8a002819 msg="SendEvent: Pushed event to core data"
    ```

5. Next step: Consume these events in your application using the [App Functions SDK](https://github.com/edgexfoundry/app-functions-sdk-go).