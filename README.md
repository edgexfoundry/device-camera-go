# Device-Camera-Go

This repository contains a camera device service compatible with EdgeX Foundry. The project utilizes [EdgeX Device SDK Go](https://github.com/edgexfoundry/device-sdk-go) to communicate with dependent EdgeX Go services. This means the service relies on the appropriate EdgeX services to be running in order to start up and function properly (Consul, Core Data, Core Metadata and others).

Use Device-Camera-Go to manage interactions between your EdgeX applications and IP network cameras. This includes the ability to capture still images or RTP video streams from ONVIF compliant IP cameras connected to your network. It also supports accessing Axis cameras, making available a similar but different interface to perform functions with the camera.

The project is intended to provide both functionality to serve as a springboard for visual application development at the edge, as well as a reference example of how to implement device services using the EdgeX Device SDK Go package.

This service has been tested with the following configuration:

IP cameras:

- 3x AXIS M3046-V Network Camera with firmware version: 6.15.6
- 1x HIKVISION Camera with firmware version: V5.4.1 build 160525

Development System:

- Oracle VirtualBox VM running Ubuntu 16.04 with 2 cores, 6GB RAM and 256MB video memory.

# Getting Started

The instructions below are the minimal steps needed to get you up and going quickly.

Additional details and answers related to EdgeX itself are available on:

- [EdgeX Wiki](https://docs.edgexfoundry.org/Ch-GettingStartedUsers.html)

# Prerequisites - Developer Setup

Perform the following steps to set up your development system. These are necessary to build and extend the device-camera-go component.

Further detailed information is available on [EdgeX-Go repository](https://github.com/edgexfoundry/edgex-go/blob/master/docs/getting-started/Ch-GettingStartedGoDevelopers.rst).

1. Install **Go 1.10** and Glide (package manager) using instructions found here:

   - https://github.com/golang/go/wiki/Ubuntu

   - NOTE: If upgrading from an earlier version of Go, it is best to first remove it; ref: https://golang.org/doc/install

   - Install **dep** with:

     ```
     $ curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
     ```

2. Install **Docker**.

   - If upgrading from an earlier version of Docker, BKM is to first remove it using:

     ```
     $ sudo apt-get remove docker docker-engine docker.io
     ```

   - Follow instructions from [Install Docker using repository](https://docs.docker.com/install/linux/docker-ce/ubuntu/#install-using-the-repository)

3. (optional) Install an **IDE** as appropriate. **[JetBrains Go Land](https://www.jetbrains.com/go/)** is a popular choice, alternate options below.

   ```
   $ tar -xzf goland-2018.2.1.tar.gz
   $ ./bin/goland.sh
   ```

   - Alternate: [Visual Studio](https://code.visualstudio.com/)
   - Alternate: [Sublime Text](https://www.alexedwards.net/blog/streamline-your-sublime-text-and-go-workflow)
   - ...

# Preparing EdgeX Stack

1. The easiest way to launch and explore device-camera-go using existing EdgeX services is to clone the edgexfoundry/developer-scripts repository and run the published Docker containers for the **EdgeX Delhi 0.7.1 release**.

   ```
   $ cd ~/dev
   $ git clone https://github.com/edgexfoundry/developer-scripts.git
   $ cd ~/dev/developer-scripts/compose-files
   ```

   At time of writing, you will see that the latest docker-compose.yml is equivalent to Delhi 0.7.1.

2. The device-camera-go device service does not require other EdgeX device services. To reduce runtime footprint a bit, and possible confusion about new devices, comment out the device-virtual section (using # at beginning of the line) in **./compose-files/docker-compose.yml**. 

   ```
   # device-virtual:
   #   image: edgexfoundry/docker-device-virtual:0.6.0
   #   ports:device-virtual:
   ...
   ```

3. If you are already running MongoDB on its default port (27017), you will want to also update **./compose-files/docker-compose.yml** to assign a unique port for EdgeX MongoDB (e.g., 27018 as the port referenced by your local system):

   ```
   mongo:
     ...
     ports:
      - "27018:27017"
   ```

4. The first time launching EdgeX, Docker will automatically **pull the EdgeX images** to your system. This requires an Internet connection and allowance through proxies/firewall. Detailed information about pulling EdgeX containers is available here: https://docs.edgexfoundry.org/Ch-GettingStartedUsersNexus.html

# Preparing Device-SDK-Go

1. Get latest device-camera-go source code:

   ```
   cd $GOPATH/src/github.com/edgexfoundry-holding
   git clone https://github.com/edgexfoundry-holding/device-camera-go.git
   ```

   Note: You may alternately clone to a separate project folder so long as you create a symlink to it from $GOPATH/src/github.com/edgexfoundry-holding folder.

2. In GoLand, load device-camera-go project, execute the following within View/Tools/Terminal:

   ```
   $ make prepare
   $ make build
   ```

   Alternately, invoke these same commands in a terminal from the root folder of the project.

# Running device-camera-go

Ready-set-go!

1. **Launch the EdgeX stack** by instructing Docker to use developer-scripts/compose-files/docker-compose.yml to run the services in detached state. This way control will return to your terminal window. Of course you can leave off the -d parameter and open additional terminals if you prefer watching EdgeX-specific logging activities in real time.

   ```
   $ cd ~/dev/developer-scripts/compose-files
   $ docker-compose up -d
   ```

2. **Confirm EdgeX services** are running properly by navigating to the Consul dashboard in your browser:

   - http://127.0.0.1:8500/ui/#/dc1/services

     [image of green/passing services]

3. **Modify your /etc/hosts file** to override DNS resolution to resolve EdgeX service URLs. Add a single line below your existing localhost entry: 

   > 127.0.0.1	localhost
   > 127.0.0.1	edgex-core-command edgex-core-metadata edgex-core-data
   >
   > ...

   Alternately, adjust EdgeX service URLs used in this guide and the tests/postman import file so they point to  "127.0.0.1", "localhost".

4. **Modify the ./run.sh script** with appropriate options (described below) and then launch the device-camera-go service.
   NOTE: Be sure to update the parameters you supply to device-camera-go with expected sources and IP address ranges according to your network topology and IP camera deployment and configuration:

   ```
   $ cd $GOPATH/src/github.com/edgexfoundry-holding/device-camera-go
   $ make run
   $ ./run
   ```

# Device-Camera-Go Configuration

## Camera Discovery

- The released image can be pulled from hub.docker.intel.com/context/saf-video-camera-discovery:v0.1.0
- Camera discovery has the following flags when running

| Parameter          | Default                    | Description                                                  |
| ------------------ | -------------------------- | ------------------------------------------------------------ |
| -registry / -r     | false                      | If this flag is set, the device-camera-go service is registered with and monitored by Consul |
| -confdir / -c      | ./res                      | Allows you to specify an alternate configuration directory. The path you specify after this parameter will be used to locate the configuration.toml and related files for loading your specified deploy-time set of camera devices. |
| -ip                | 192.168.0.1-30             | IP address/range for nmap scan.<br />                        |
| -mask              | ""                         | Network mask for nmap scan<br />Subnet scan example:<br/>-ip "10.25.35.*" - mask "/24"<br /><br />Single IP example:<br />-ip = "10.1.24.101" -mask = "32" -scanduration = "10s"<br /> |
| -scanduration      | "15s"                      | Controls how long to allow IP groups to be scanned before timing out. The value may specify units of seconds, minutes, or hours ("s", "m", "h", respectively) |
| -interval          | 60                         | Frequency of scans to schedule. Units are in seconds. The lower the number the more responsive, but greater network overhead and system I/O required. At a minimum, the value must be 10 seconds longer than the configured scan duration but for practical purposes should be set to much larger values to allow IP groups to scan (highly relative to # IP and # ports being scanned). |
| -source  ...       | (see description)          | Source to scan. Provides a way to segment whether ONVIF and/or Axis cameras should be discovered. When no colon/port is provided, default port is 554 for Axis and 80 for ONVIF.<br />Examples:<br />-source axis<br />-source onvif<br />-source axis -source onvif<br />-source axis:554 -source onvif:80<br />-source axis:554,557,558 -source onvif:80,8000,8020 |
| -tagsFile          | ./res/tags.json            | Path to file containing camera tags. These are used on startup and will reflect any changes occurring by invoking /tags command. |
| -cameracredentials | ./res/cam-credentials.conf | Location and name of the file where camera access credentials Path to file containing credentials for ONVIF cameras. This holds a user and password, tab-separated that is configured on the camera for purpose of being accessed by the device-camera-go service. |

```
docker run --name device-camera-go -it --network host --rm hub.docker.com/device-camera-go:v0.1.0 <parameters>
```

```
$ go run main.go -registry -source onvif -source axis
```

## Build - and other Make Targets

| Makefile Command | Action Performed                                             |
| ---------------- | ------------------------------------------------------------ |
| make prepare     | Initializes the project and dependencies                     |
| make update      | Updates components                                           |
| make build       | Builds camera-device-go binary                               |
| make run         | Converts run.sh content into ./run launcher                  |
| make test        | Invokes gometalinter (after installing for you if needed) and go test |
| make clean       | Removes camera-device-go binary                              |

## Tests

A set of handy Postman links are available in the **./tests** subfolder.

Since name prefixes are prepended to camera serialnumbers to uniquely identify (and namespace) your EdgeX IP camera devices. This constructs an immutable way to access a discovered IP camera; the name will not change across service start/stop or even removing information from the backend database. As a result, you will find it useful to add Postman requests using these device names to verify results. 

For example, replace the serial number ACCC8E8439F0 with your camera's serial number on this URI:

- http://edgex-core-command:48082/api/v1/device/name/edgex-camera-onvif-ACCC8E8439F0

## Persisted Resources

File-based resources are kept in the **./res** folder. This can be configured with the -confdir parameter to point to a different location.  and cache backing stores and stored in ./res folder.

## Configuration.toml

The ./res/configuration.toml will dynamically define and register devices based on the information provided. It links to the described EdgeX device profile (constructing virtual Axis camera devices) as a template to populate with these static cameras known at deploy-time.

## Tags File

An example of values to populate camera tags is found in the ./res folder. Note that the keys are simply the device serial number. So when the associated device is discovered using both ONVIF and Axis, you will find a separate device name registered with EdgeX for each device profile (uniquely named per a name prefix).

> {
> "ACCC8E8439F0":{"friendly_name":"Black Axis","location":"North Wall","newtag":"sometag","store":true},
> "ACCC8E843A18":{"friendly_name":"Black Axis2","location":"North Wall2","newtag":"sometag2","store":true},
> "ACCC8E8621BB":{"friendly_name":"Black Axis3","location":"Ceiling","newtag":"sometag3","store":false},
> "DS-2CD2342WD-I20160817BBWR634390011":{"friendly_name":"White HikVision","location":"Wall3","newtag":"sometag4","store":true}
> }

## Commands

Identify the available commands from EdgeX core command using your device name:

http://edgex-core-command:48082/api/v1/device/name/edgex-camera-onvif-ACCC8E8439F0

This reveals the commands that were registered for this camera; e.g., **tags** and **onvif_profiles**.

> {
>  "id": "5c513b9a9f8fc20001a711aa",
>  "name": "edgex-camera-onvif-ACCC8E8439F0",
>  "adminState": "UNLOCKED",
>  "operatingState": "ENABLED",
>  "lastConnected": 0,
>  "lastReported": 0,
>  "labels": [
>      "newtag:sometag",
>      "store:true",
>      "friendly_name:Black Axis",
>      "location:North Wall"
>  ],
>  "location": null,
>  "commands": [
>      {
>          "created": 1548827534765,
>          "modified": 0,
>          "origin": 0,
>          "id": "5c513b8e9f8fc20001a711a0",
>          "name": "tags",
>          "get": {
>              "path": "/api/v1/device/{deviceId}/tags",
>              "responses": [
>                  {
>                      "code": "200",
>                      "description": "Get Tags",
>                      "expectedValues": [
>                          "cameradevice_tags"
>                      ]
>                  },
>                  {
>                      "code": "503",
>                      "description": "Get Tags Error",
>                      "expectedValues": [
>                          "cameradevice_error"
>                      ]
>                  }
>              ],
>              "url": "http://edgex-core-command:48082/api/v1/device/5c513b9a9f8fc20001a711aa/command/5c513b8e9f8fc20001a711a0"
>          },
>          "put": {
>              "path": "/api/v1/device/{deviceId}/tags",
>              "responses": [
>                  {
>                      "code": "200",
>                      "description": "Set Tags",
>                      "expectedValues": [
>                          "cameradevice_id"
>                      ]
>                  },
>                  {
>                      "code": "503",
>                      "description": "Set Tags Error",
>                      "expectedValues": [
>                          "cameradevice_error"
>                      ]
>                  }
>              ],
>              "parameterNames": [
>                  "cameradevice_tags"
>              ],
>              "url": "http://edgex-core-command:48082/api/v1/device/5c513b9a9f8fc20001a711aa/command/5c513b8e9f8fc20001a711a0"
>          }
>      },
>      {
>          "created": 1548827534765,
>          "modified": 0,
>          "origin": 0,
>          "id": "5c513b8e9f8fc20001a711a1",
>          "name": "onvif_profiles",
>          "get": {
>              "path": "/api/v1/device/{deviceId}/onvif_profiles",
>              "responses": [
>                  {
>                      "code": "200",
>                      "description": "Get ONVIF Profiles",
>                      "expectedValues": [
>                          "onvif_camera_metadata"
>                      ]
>                  },
>                  {
>                      "code": "503",
>                      "description": "Get ONVIF Profiles Error",
>                      "expectedValues": [
>                          "cameradevice_error"
>                      ]
>                  }
>              ],
>              "url": "http://edgex-core-command:48082/api/v1/device/5c513b9a9f8fc20001a711aa/command/5c513b8e9f8fc20001a711a1"
>          },
>          "put": {
>              "path": "/api/v1/device/{deviceId}/onvif_profiles",
>              "responses": [
>                  {
>                      "code": "200",
>                      "description": "Set ONVIF Profiles",
>                      "expectedValues": [
>                          "onvif_camera_metadata"
>                      ]
>                  },
>                  {
>                      "code": "503",
>                      "description": "Set ONVIF Profiles Error",
>                      "expectedValues": [
>                          "cameradevice_error"
>                      ]
>                  }
>              ],
>              "parameterNames": [
>                  "onvif_camera_metadata"
>              ],
>              "url": "http://edgex-core-command:48082/api/v1/device/5c513b9a9f8fc20001a711aa/command/5c513b8e9f8fc20001a711a1"
>          }
>      }
>  ]
> }

Note that commands are initiated through the EdgeX Command Service through the described "url" properties that use derived values (e.g., the unique device id and command id values) which are constructed by EdgeX at runtime.

### Tags Command

Invoking a **tags** command on our ONVIF camera device with serial number ACCC8E8439F0 responds as follows. Note that commands are initiated through the EdgeX Command Service with derived values at runtime.

> {
>  "id": "",
>  "pushed": 0,
>  "device": "edgex-camera-onvif-ACCC8E8439F0",
>  "created": 0,
>  "modified": 0,
>  "origin": 1548867163784,
>  "schedule": null,
>  "event": null,
>  "readings": [
>      {
>          "id": "",
>          "pushed": 0,
>          "created": 0,
>          "origin": 1548867163784,
>          "modified": 0,
>          "device": "edgex-camera-onvif-ACCC8E8439F0",
>          "name": "cameradevice_tags",
>          "value": "'friendly_name':'Black Axis','location':'North Wall','newtag':'sometag','store':true"
>      }
>  ]
> }

### ONVIF Profiles Command

Invoking the **onvif_profiles** command will return device readings with a value that holds camera connectivity information similar to the following:

> {
> 	"id": "",
> 	"device": "edgex-camera-onvif-ACCC8E8439F0",
> 	"origin": 1548868440409,  ...
> 	"readings": [{
> 		"origin": 1548868440409,
> 		"device": "edgex-camera-onvif-ACCC8E8439F0",
> 		"name": "onvif_camera_metadata",
> 		"value": "{\"ip\":\"10.43.18.158:80\",\"productname\":\"AXIS\",\"firmwareversion\":\"6.15.6\",\"serialnumber\":\"ACCC8E8439F0\",\"profiles\":[{\"ProfileName\":\"profile_1 h264\",\"Formats\":\"H264\",\"Resolutions\":[\"1920\",\"1080\"],\"RTSPPath\":\"rtsp://10.43.18.158/onvif-media/media.amp?profile=profile_1_h264\\u0026sessiontimeout=60\\u0026streamtype=unicast\",\"ImagePath\":\"http://10.43.18.158/onvif-cgi/jpg/image.cgi?resolution=1920x1080\\u0026compression=30\",\"ProfileToken\":\"profile_1_h264\"},{\"ProfileName\":\"profile_1 jpeg\",\"Formats\":\"JPEG\",\"Resolutions\":[\"1920\",\"1080\"],\"RTSPPath\":\"rtsp://10.43.18.158/onvif-media/media.amp?profile=profile_1_jpeg\\u0026sessiontimeout=60\\u0026streamtype=unicast\",\"ImagePath\":\"http://10.43.18.158/onvif-cgi/jpg/image.cgi?resolution=1920x1080\\u0026compression=30\",\"ProfileToken\":\"profile_1_jpeg\"}],\"tags\":{\"friendly_name\":\"Black Axis\",\"location\":\"North Wall\",\"newtag\":\"sometag\",\"store\":true}}"
> 	}]
> }

## Device Data Models

### **ONVIF Device Data Model**

Device response you should expect from querying EdgeX for devices by ONVIF profile. As a concrete example, with EdgeX and device-camera-go service running, invoke a GET request: 
http://edgex-core-metadata:48081/api/v1/device/profilename/camera-profile-onvif

This will produce a set of known devices, whether produced at startup by modifying configuration.toml and/or discovered on the network actively at your configured interval.

These responses reveal names, associated device profile, commands and other metadata associated with the device.

Example Response:

> [
>  {
>      "created": 1548827546764,
>      "modified": 1548827546764,
>      "origin": 1548827546759,
>      "description": "",
>      "id": "5c513b9a9f8fc20001a711aa",
>      "name": "edgex-camera-onvif-ACCC8E8439F0",
>      "adminState": "UNLOCKED",
>      "operatingState": "ENABLED",
>      "addressable": {
>          "created": 1548827546759,
>          "modified": 0,
>          "origin": 1548827546758,
>          "id": "5c513b9a9f8fc20001a711a9",
>          "name": "edgex-camera-onvif-ACCC8E8439F0",
>          "protocol": "HTTP",
>          "method": null,
>          "address": "172.17.0.1",
>          "port": 49990,
>          "path": "/cameradiscoveryprovider",
>          "publisher": "none",
>          "user": "none",
>          "password": "none",
>          "topic": "none",
>          "baseURL": "HTTP://172.17.0.1:49990",
>          "url": "HTTP://172.17.0.1:49990/cameradiscoveryprovider"
>      },
>      "lastConnected": 0,
>      "lastReported": 0,
>      "labels": [
>          "newtag:sometag",
>          "store:true",
>          "friendly_name:Black Axis",
>          "location:North Wall"
>      ],
>      "location": null,
>      "service": {
>          "created": 1548827534696,
>          "modified": 1548827534696,
>          "origin": 1548827534695,
>          "description": "",
>          "id": "5c513b8e9f8fc20001a7119c",
>          "name": "device-camera-go",
>          "lastConnected": 0,
>          "lastReported": 0,
>          "operatingState": "ENABLED",
>          "labels": [],
>          "addressable": {
>              "created": 1548827534692,
>              "modified": 0,
>              "origin": 1548827534689,
>              "id": "5c513b8e9f8fc20001a7119b",
>              "name": "device-camera-go",
>              "protocol": "HTTP",
>              "method": "POST",
>              "address": "172.17.0.1",
>              "port": 49990,
>              "path": "/api/v1/callback",
>              "publisher": null,
>              "user": null,
>              "password": null,
>              "topic": null,
>              "baseURL": "HTTP://172.17.0.1:49990",
>              "url": "HTTP://172.17.0.1:49990/api/v1/callback"
>          },
>          "adminState": "UNLOCKED"
>      },
>      "profile": {
>          "created": 1548827534765,
>          "modified": 1548827534765,
>          "origin": 0,
>          "description": "EdgeX device profile for ONVIF supports conformant IP cameras.",
>          "id": "5c513b8e9f8fc20001a711a2",
>          "name": "camera-profile-onvif",
>          "manufacturer": "www.onvif.org",
>          "model": "EdgeX_CameraDevice",
>          "labels": [
>              "camera-onvif"
>          ],
>          "objects": null,
>          "deviceResources": [
>              {
>                  "description": "command error response",
>                  "name": "cam_error",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "An Error Occurred",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "CamErrorString"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device tags",
>                  "name": "tags",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "Tags"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device standards-based metadata",
>                  "name": "onvif_profiles",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "ONVIFProfiles"
>                      }
>                  },
>                  "attributes": null
>              }
>          ],
>          "resources": [
>              {
>                  "name": "cam_error",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "onvif_profiles",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              }
>          ],
>          "commands": [
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a0",
>                  "name": "tags",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get Tags",
>                              "expectedValues": [
>                                  "cameradevice_tags"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set Tags",
>                              "expectedValues": [
>                                  "cameradevice_id"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "cameradevice_tags"
>                      ]
>                  }
>              },
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a1",
>                  "name": "onvif_profiles",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "onvif_camera_metadata"
>                      ]
>                  }
>              }
>          ]
>      }
>  },
>  {
>      "created": 1548827546836,
>      "modified": 1548827546836,
>      "origin": 1548827546803,
>      "description": "",
>      "id": "5c513b9a9f8fc20001a711ac",
>      "name": "edgex-camera-onvif-ACCC8E8621BB",
>      "adminState": "UNLOCKED",
>      "operatingState": "ENABLED",
>      "addressable": {
>          "created": 1548827546803,
>          "modified": 0,
>          "origin": 1548827546802,
>          "id": "5c513b9a9f8fc20001a711ab",
>          "name": "edgex-camera-onvif-ACCC8E8621BB",
>          "protocol": "HTTP",
>          "method": null,
>          "address": "172.17.0.1",
>          "port": 49990,
>          "path": "/cameradiscoveryprovider",
>          "publisher": "none",
>          "user": "none",
>          "password": "none",
>          "topic": "none",
>          "baseURL": "HTTP://172.17.0.1:49990",
>          "url": "HTTP://172.17.0.1:49990/cameradiscoveryprovider"
>      },
>      "lastConnected": 0,
>      "lastReported": 0,
>      "labels": [
>          "friendly_name:Black Axis3",
>          "location:Ceiling",
>          "newtag:sometag3",
>          "store:false"
>      ],
>      "location": null,
>      "service": {
>          "created": 1548827534696,
>          "modified": 1548827534696,
>          "origin": 1548827534695,
>          "description": "",
>          "id": "5c513b8e9f8fc20001a7119c",
>          "name": "device-camera-go",
>          "lastConnected": 0,
>          "lastReported": 0,
>          "operatingState": "ENABLED",
>          "labels": [],
>          "addressable": {
>              "created": 1548827534692,
>              "modified": 0,
>              "origin": 1548827534689,
>              "id": "5c513b8e9f8fc20001a7119b",
>              "name": "device-camera-go",
>              "protocol": "HTTP",
>              "method": "POST",
>              "address": "172.17.0.1",
>              "port": 49990,
>              "path": "/api/v1/callback",
>              "publisher": null,
>              "user": null,
>              "password": null,
>              "topic": null,
>              "baseURL": "HTTP://172.17.0.1:49990",
>              "url": "HTTP://172.17.0.1:49990/api/v1/callback"
>          },
>          "adminState": "UNLOCKED"
>      },
>      "profile": {
>          "created": 1548827534765,
>          "modified": 1548827534765,
>          "origin": 0,
>          "description": "EdgeX device profile for ONVIF supports conformant IP cameras.",
>          "id": "5c513b8e9f8fc20001a711a2",
>          "name": "camera-profile-onvif",
>          "manufacturer": "www.onvif.org",
>          "model": "EdgeX_CameraDevice",
>          "labels": [
>              "camera-onvif"
>          ],
>          "objects": null,
>          "deviceResources": [
>              {
>                  "description": "command error response",
>                  "name": "cam_error",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "An Error Occurred",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "CamErrorString"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device tags",
>                  "name": "tags",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "Tags"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device standards-based metadata",
>                  "name": "onvif_profiles",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "ONVIFProfiles"
>                      }
>                  },
>                  "attributes": null
>              }
>          ],
>          "resources": [
>              {
>                  "name": "cam_error",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "onvif_profiles",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              }
>          ],
>          "commands": [
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a0",
>                  "name": "tags",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get Tags",
>                              "expectedValues": [
>                                  "cameradevice_tags"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set Tags",
>                              "expectedValues": [
>                                  "cameradevice_id"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "cameradevice_tags"
>                      ]
>                  }
>              },
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a1",
>                  "name": "onvif_profiles",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "onvif_camera_metadata"
>                      ]
>                  }
>              }
>          ]
>      }
>  },
>  {
>      "created": 1548827546856,
>      "modified": 1548827546856,
>      "origin": 1548827546849,
>      "description": "",
>      "id": "5c513b9a9f8fc20001a711ae",
>      "name": "edgex-camera-onvif-DS-2CD2342WD-I20160817BBWR634390011",
>      "adminState": "UNLOCKED",
>      "operatingState": "ENABLED",
>      "addressable": {
>          "created": 1548827546848,
>          "modified": 0,
>          "origin": 1548827546842,
>          "id": "5c513b9a9f8fc20001a711ad",
>          "name": "edgex-camera-onvif-DS-2CD2342WD-I20160817BBWR634390011",
>          "protocol": "HTTP",
>          "method": null,
>          "address": "172.17.0.1",
>          "port": 49990,
>          "path": "/cameradiscoveryprovider",
>          "publisher": "none",
>          "user": "none",
>          "password": "none",
>          "topic": "none",
>          "baseURL": "HTTP://172.17.0.1:49990",
>          "url": "HTTP://172.17.0.1:49990/cameradiscoveryprovider"
>      },
>      "lastConnected": 0,
>      "lastReported": 0,
>      "labels": [
>          "newtag:sometag4",
>          "store:true",
>          "friendly_name:White HikVision",
>          "location:Wall3"
>      ],
>      "location": null,
>      "service": {
>          "created": 1548827534696,
>          "modified": 1548827534696,
>          "origin": 1548827534695,
>          "description": "",
>          "id": "5c513b8e9f8fc20001a7119c",
>          "name": "device-camera-go",
>          "lastConnected": 0,
>          "lastReported": 0,
>          "operatingState": "ENABLED",
>          "labels": [],
>          "addressable": {
>              "created": 1548827534692,
>              "modified": 0,
>              "origin": 1548827534689,
>              "id": "5c513b8e9f8fc20001a7119b",
>              "name": "device-camera-go",
>              "protocol": "HTTP",
>              "method": "POST",
>              "address": "172.17.0.1",
>              "port": 49990,
>              "path": "/api/v1/callback",
>              "publisher": null,
>              "user": null,
>              "password": null,
>              "topic": null,
>              "baseURL": "HTTP://172.17.0.1:49990",
>              "url": "HTTP://172.17.0.1:49990/api/v1/callback"
>          },
>          "adminState": "UNLOCKED"
>      },
>      "profile": {
>          "created": 1548827534765,
>          "modified": 1548827534765,
>          "origin": 0,
>          "description": "EdgeX device profile for ONVIF supports conformant IP cameras.",
>          "id": "5c513b8e9f8fc20001a711a2",
>          "name": "camera-profile-onvif",
>          "manufacturer": "www.onvif.org",
>          "model": "EdgeX_CameraDevice",
>          "labels": [
>              "camera-onvif"
>          ],
>          "objects": null,
>          "deviceResources": [
>              {
>                  "description": "command error response",
>                  "name": "cam_error",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "An Error Occurred",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "CamErrorString"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device tags",
>                  "name": "tags",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "Tags"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device standards-based metadata",
>                  "name": "onvif_profiles",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "ONVIFProfiles"
>                      }
>                  },
>                  "attributes": null
>              }
>          ],
>          "resources": [
>              {
>                  "name": "cam_error",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "onvif_profiles",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              }
>          ],
>          "commands": [
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a0",
>                  "name": "tags",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get Tags",
>                              "expectedValues": [
>                                  "cameradevice_tags"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set Tags",
>                              "expectedValues": [
>                                  "cameradevice_id"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "cameradevice_tags"
>                      ]
>                  }
>              },
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a1",
>                  "name": "onvif_profiles",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "onvif_camera_metadata"
>                      ]
>                  }
>              }
>          ]
>      }
>  },
>  {
>      "created": 1548827546871,
>      "modified": 1548827546871,
>      "origin": 1548827546863,
>      "description": "",
>      "id": "5c513b9a9f8fc20001a711b0",
>      "name": "edgex-camera-onvif-ACCC8E843A18",
>      "adminState": "UNLOCKED",
>      "operatingState": "ENABLED",
>      "addressable": {
>          "created": 1548827546862,
>          "modified": 0,
>          "origin": 1548827546860,
>          "id": "5c513b9a9f8fc20001a711af",
>          "name": "edgex-camera-onvif-ACCC8E843A18",
>          "protocol": "HTTP",
>          "method": null,
>          "address": "172.17.0.1",
>          "port": 49990,
>          "path": "/cameradiscoveryprovider",
>          "publisher": "none",
>          "user": "none",
>          "password": "none",
>          "topic": "none",
>          "baseURL": "HTTP://172.17.0.1:49990",
>          "url": "HTTP://172.17.0.1:49990/cameradiscoveryprovider"
>      },
>      "lastConnected": 0,
>      "lastReported": 0,
>      "labels": [
>          "newtag:sometag2",
>          "store:true",
>          "friendly_name:Black Axis2",
>          "location:North Wall2"
>      ],
>      "location": null,
>      "service": {
>          "created": 1548827534696,
>          "modified": 1548827534696,
>          "origin": 1548827534695,
>          "description": "",
>          "id": "5c513b8e9f8fc20001a7119c",
>          "name": "device-camera-go",
>          "lastConnected": 0,
>          "lastReported": 0,
>          "operatingState": "ENABLED",
>          "labels": [],
>          "addressable": {
>              "created": 1548827534692,
>              "modified": 0,
>              "origin": 1548827534689,
>              "id": "5c513b8e9f8fc20001a7119b",
>              "name": "device-camera-go",
>              "protocol": "HTTP",
>              "method": "POST",
>              "address": "172.17.0.1",
>              "port": 49990,
>              "path": "/api/v1/callback",
>              "publisher": null,
>              "user": null,
>              "password": null,
>              "topic": null,
>              "baseURL": "HTTP://172.17.0.1:49990",
>              "url": "HTTP://172.17.0.1:49990/api/v1/callback"
>          },
>          "adminState": "UNLOCKED"
>      },
>      "profile": {
>          "created": 1548827534765,
>          "modified": 1548827534765,
>          "origin": 0,
>          "description": "EdgeX device profile for ONVIF supports conformant IP cameras.",
>          "id": "5c513b8e9f8fc20001a711a2",
>          "name": "camera-profile-onvif",
>          "manufacturer": "www.onvif.org",
>          "model": "EdgeX_CameraDevice",
>          "labels": [
>              "camera-onvif"
>          ],
>          "objects": null,
>          "deviceResources": [
>              {
>                  "description": "command error response",
>                  "name": "cam_error",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "An Error Occurred",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "CamErrorString"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device tags",
>                  "name": "tags",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "Tags"
>                      }
>                  },
>                  "attributes": null
>              },
>              {
>                  "description": "camera device standards-based metadata",
>                  "name": "onvif_profiles",
>                  "tag": null,
>                  "properties": {
>                      "value": {
>                          "type": "String",
>                          "readWrite": "RW",
>                          "minimum": null,
>                          "maximum": null,
>                          "defaultValue": "key:value,key:value",
>                          "size": null,
>                          "word": "2",
>                          "lsb": null,
>                          "mask": "0x00",
>                          "shift": "0",
>                          "scale": "1.0",
>                          "offset": "0.0",
>                          "base": "0",
>                          "assertion": null,
>                          "signed": true,
>                          "precision": null
>                      },
>                      "units": {
>                          "type": "String",
>                          "readWrite": "R",
>                          "defaultValue": "ONVIFProfiles"
>                      }
>                  },
>                  "attributes": null
>              }
>          ],
>          "resources": [
>              {
>                  "name": "cam_error",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "cam_error",
>                          "property": "value",
>                          "parameter": "cameradevice_error",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "tags",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "tags",
>                          "property": "value",
>                          "parameter": "cameradevice_tags",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              },
>              {
>                  "name": "onvif_profiles",
>                  "get": [
>                      {
>                          "index": null,
>                          "operation": "get",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ],
>                  "set": [
>                      {
>                          "index": null,
>                          "operation": "set",
>                          "object": "onvif_profiles",
>                          "property": "value",
>                          "parameter": "onvif_camera_metadata",
>                          "resource": null,
>                          "secondary": [],
>                          "mappings": {}
>                      }
>                  ]
>              }
>          ],
>          "commands": [
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a0",
>                  "name": "tags",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get Tags",
>                              "expectedValues": [
>                                  "cameradevice_tags"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/tags",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set Tags",
>                              "expectedValues": [
>                                  "cameradevice_id"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set Tags Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "cameradevice_tags"
>                      ]
>                  }
>              },
>              {
>                  "created": 1548827534765,
>                  "modified": 0,
>                  "origin": 0,
>                  "id": "5c513b8e9f8fc20001a711a1",
>                  "name": "onvif_profiles",
>                  "get": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Get ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Get ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ]
>                  },
>                  "put": {
>                      "path": "/api/v1/device/{deviceId}/onvif_profiles",
>                      "responses": [
>                          {
>                              "code": "200",
>                              "description": "Set ONVIF Profiles",
>                              "expectedValues": [
>                                  "onvif_camera_metadata"
>                              ]
>                          },
>                          {
>                              "code": "503",
>                              "description": "Set ONVIF Profiles Error",
>                              "expectedValues": [
>                                  "cameradevice_error"
>                              ]
>                          }
>                      ],
>                      "parameterNames": [
>                          "onvif_camera_metadata"
>                      ]
>                  }
>              }
>          ]
>      }
>  }
> ]

### ONVIF Metadata Cache Examples

Example response from results that you can expect to find populates the locally cached **ONVIF metadata** ./res/camerainfo.json

This serves as an in-memory and persisted repository to concurrently provide responses to commands such as **GET /api/v1/device/{deviceId}/onvif_profiles**

> [{
> 	"ip": "10.43.18.149:80",
> 	"productname": "HIKVISION",
> 	"firmwareversion": "V5.4.1 build 160525",
> 	"serialnumber": "DS-2CD2342WD-I20160817BBWR634390011",
> 	"profiles": [{
> 		"ProfileName": "mainStream",
> 		"Formats": "H264",
> 		"Resolutions": ["1280", "720"],
> 		"RTSPPath": "rtsp://10.43.18.149:554/Streaming/Channels/101?transportmode=unicast\u0026profile=Profile_1",
> 		"ImagePath": "http://10.43.18.149/onvif-http/snapshot?Profile_1",
> 		"ProfileToken": "Profile_1"
> 	}, {
> 		"ProfileName": "subStream",
> 		"Formats": "H264",
> 		"Resolutions": ["640", "360"],
> 		"RTSPPath": "rtsp://10.43.18.149:554/Streaming/Channels/102?transportmode=unicast\u0026profile=Profile_2",
> 		"ImagePath": "http://10.43.18.149/onvif-http/snapshot?Profile_2",
> 		"ProfileToken": "Profile_2"
> 	}],
> 	"tags": {
> 		"friendly_name": "White HikVision",
> 		"location": "Wall3",
> 		"newtag": "sometag4",
> 		"store": true
> 	}
> }, {
> 	"ip": "10.43.18.170:80",
> 	"productname": "AXIS",
> 	"firmwareversion": "7.15.2.1",
> 	"serialnumber": "ACCC8E8621BB",
> 	"profiles": [{
> 		"ProfileName": "profile_0 h264",
> 		"Formats": "H264",
> 		"Resolutions": ["1920", "1920"],
> 		"RTSPPath": "rtsp://10.43.18.170/onvif-media/media.amp?profile=profile_0_h264\u0026sessiontimeout=60\u0026streamtype=unicast",
> 		"ImagePath": "http://10.43.18.170/onvif-cgi/jpg/image.cgi?resolution=1920x1920\u0026compression=30",
> 		"ProfileToken": "profile_0_h264"
> 	}, {
> 		"ProfileName": "profile_1 h264",
> 		"Formats": "H264",
> 		"Resolutions": ["1920", "1920"],
> 		"RTSPPath": "rtsp://10.43.18.170/onvif-media/media.amp?profile=profile_1_h264\u0026sessiontimeout=60\u0026streamtype=unicast",
> 		"ImagePath": "http://10.43.18.170/onvif-cgi/jpg/image.cgi?resolution=1920x1920\u0026compression=30\u0026camera=2",
> 		"ProfileToken": "profile_1_h264"
> 	}, {
> 		"ProfileName": "profile_0 jpeg",
> 		"Formats": "JPEG",
> 		"Resolutions": ["1920", "1920"],
> 		"RTSPPath": "rtsp://10.43.18.170/onvif-media/media.amp?profile=profile_0_jpeg\u0026sessiontimeout=60\u0026streamtype=unicast",
> 		"ImagePath": "http://10.43.18.170/onvif-cgi/jpg/image.cgi?resolution=1920x1920\u0026compression=30",
> 		"ProfileToken": "profile_0_jpeg"
> 	}, {
> 		"ProfileName": "profile_1 jpeg",
> 		"Formats": "JPEG",
> 		"Resolutions": ["1920", "1920"],
> 		"RTSPPath": "rtsp://10.43.18.170/onvif-media/media.amp?profile=profile_1_jpeg\u0026sessiontimeout=60\u0026streamtype=unicast",
> 		"ImagePath": "http://10.43.18.170/onvif-cgi/jpg/image.cgi?resolution=1920x1920\u0026compression=30\u0026camera=2",
> 		"ProfileToken": "profile_1_jpeg"
> 	}],
> 	"tags": {
> 		"friendly_name": "Black Axis3",
> 		"location": "Ceiling",
> 		"newtag": "sometag3",
> 		"store": false
> 	}
> }, {
> 	"ip": "10.43.18.158:80",
> 	"productname": "AXIS",
> 	"firmwareversion": "6.15.6",
> 	"serialnumber": "ACCC8E8439F0",
> 	"profiles": [{
> 		"ProfileName": "profile_1 h264",
> 		"Formats": "H264",
> 		"Resolutions": ["1920", "1080"],
> 		"RTSPPath": "rtsp://10.43.18.158/onvif-media/media.amp?profile=profile_1_h264\u0026sessiontimeout=60\u0026streamtype=unicast",
> 		"ImagePath": "http://10.43.18.158/onvif-cgi/jpg/image.cgi?resolution=1920x1080\u0026compression=30",
> 		"ProfileToken": "profile_1_h264"
> 	}, {
> 		"ProfileName": "profile_1 jpeg",
> 		"Formats": "JPEG",
> 		"Resolutions": ["1920", "1080"],
> 		"RTSPPath": "rtsp://10.43.18.158/onvif-media/media.amp?profile=profile_1_jpeg\u0026sessiontimeout=60\u0026streamtype=unicast",
> 		"ImagePath": "http://10.43.18.158/onvif-cgi/jpg/image.cgi?resolution=1920x1080\u0026compression=30",
> 		"ProfileToken": "profile_1_jpeg"
> 	}],
> 	"tags": {
> 		"friendly_name": "Black Axis",
> 		"location": "North Wall",
> 		"newtag": "sometag",
> 		"store": true
> 	}
> }, ... ]

### **Axis Metadata Cache - Example**

Note that the data model response for **Axis camera metadata** provides similar values but these differ from the use of ONVIF profiles and other elements so it provides a different method for interacting with the devices.

[{
	"supportedresolutions": ["1280x960", "1024x768", "800x600", "640x480", "320x240", "2688x1520", "1920x1080", "1280x720", "1024x576", "640x360", "352x240"],
	"supportedformats": ["jpeg", "mjpeg", "h264"],
	"rtspauthenabled": true,
	"rtsppath": "rtsp://10.43.18.141/axis-media/media.amp",
	"imagepath": "http://10.43.18.141/jpg/1/image.jpg?resolution=352x240",
	"ip": "10.43.18.141",
	"productname": "AXIS M3046-V Network Camera",
	"firmwareversion": "6.15.6",
	"serialnumber": "ACCC8E843A18",
	"tags": {
		"friendly_name": "Black Axis2",
		"location": "North Wall2",
		"newtag": "sometag2",
		"store": true
	}
}, {
	"supportedresolutions": ["1280x960", "1024x768", "800x600", "640x480", "320x240", "2688x1520", "1920x1080", "1280x720", "1024x576", "640x360", "352x240"],
	"supportedformats": ["jpeg", "mjpeg", "h264"],
	"rtspauthenabled": true,
	"rtsppath": "rtsp://10.43.18.158/axis-media/media.amp",
	"imagepath": "http://10.43.18.158/jpg/1/image.jpg?resolution=352x240",
	"ip": "10.43.18.158",
	"productname": "AXIS M3046-V Network Camera",
	"firmwareversion": "6.15.6",
	"serialnumber": "ACCC8E8439F0",
	"tags": {
		"friendly_name": "Black Axis",
		"location": "North Wall",
		"newtag": "sometag",
		"store": true
	}
}]



## Known Issues and Improvement Opportunities

Please add as "issues" in github above ^ 



