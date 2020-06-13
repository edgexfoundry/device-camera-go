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

#### Integrate Into Applications

1. Consume these application/CBOR events in your application by cloning the [App Functions SDK](https://github.com/edgexfoundry/app-functions-sdk-go).
    ```
    git clone https://github.com/edgexfoundry/app-functions-sdk-go.git
    cd app-functions-sdk-go
    ```

2. Update line 38 of app-functions-sdk Makefile so it will build a docker image that runs 'example/simple-cbor-filter' when launched.
    ```
    ...
        -f examples/simple-cbor-filter/Dockerfile \
    ...
    ```

3. Build the application example:
    ```
    make docker
    ```

4. Update your docker-compose.yml to include your application container:
    ```
    #################################################################
    # Device Services
    #################################################################
    docker-simple-cbor-filter:
        image: edgexfoundry/docker-app-functions-sdk-go-simple:0.0.0-dev
        ports:
        - "48095:48095"
        container_name: simple-cbor-filter
        hostname: simple-cbor-filter
        networks:
        edgex-network:
            aliases:
            - simple-cbor-filter
        volumes:
        - db-data:/data/db
        - log-data:/edgex/logs
        - consul-config:/consul/config
        - consul-data:/consul/data
        depends_on:
        - data
        - command
        - metadata
        - docker-device-camera-go

    ...
    docker-device-camera-go:
        image: device-camera-go:latest
    ...
    ```

5. Launch your application service:
    ```
    docker-compose-up docker-simple-cbor-filter
    ```

6. This will show the snapshot image events being received and processed by your application. It shows log entries for the resolution of the image and the color of the pixel at the center of your camera.
    ```
    Starting edgex-device-camera-go ... done
    Creating simple-cbor-filter     ... done
    Attaching to simple-cbor-filter
    simple-cbor-filter           | Configuration pushed to registry with service key: sampleCborFilter
    simple-cbor-filter           | Configuration & Registry initializedlevel=INFO ts=2019-11-01T22:15:05.429607466Z app=sampleCborFilter source=sdk.go:357 msg="Logger successfully initialized"
    simple-cbor-filter           | level=INFO version=0.0.0 msg="Skipping core service version compatibility check for SDK Beta version or running in debugger"
    simple-cbor-filter           | level=INFO msg="Clients initialized"
    simple-cbor-filter           | level=INFO msg="Registering standard routes..."
    simple-cbor-filter           | level=INFO msg="Filtering for [onvif_snapshot] value descriptors..."
    simple-cbor-filter           | level=INFO msg="MessageBus trigger selected"
    simple-cbor-filter           | level=INFO msg="Initializing Message Bus Trigger. Subscribing to topic: events on port 5563 , Publish Topic: somewhere on port 5564"
    simple-cbor-filter           | level=INFO msg="Listening for changes from registry"
    simple-cbor-filter           | level=INFO msg="StoreAndForward disabled. Not running retry loop."
    simple-cbor-filter           | level=INFO msg="Simple CBOR Filter Application Service started"
    simple-cbor-filter           | level=INFO msg="Starting CPU Usage Average loop"
    simple-cbor-filter           | level=INFO msg="Starting HTTP Server on port :48095"
    simple-cbor-filter           | level=INFO msg="Writable configuration has been updated from Registry"
    simple-cbor-filter           | Received Image from Device: CasualWatcher001, ReadingName: onvif_snapshot, Image Type: jpeg, Image Size: (1280,720), Color in middle: {112 125 128}
    simple-cbor-filter           | Received Image from Device: CasualWatcher001, ReadingName: onvif_snapshot, Image Type: jpeg, Image Size: (1280,720), Color in middle: {114 125 128}
    ```
    