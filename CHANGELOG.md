
<a name="EdgeX Camera Device Service (found in device-camera-go) Changelog"></a>
## EdgeX Camera Device Service
[Github repository](https://github.com/edgexfoundry/device-camera-go)

## [v2.0.0] Ireland - 2021-06-30  (Not Compatible with 1.x releases)

### Features ✨
- Enable using MessageBus as the default ([#9d8893d](https://github.com/edgexfoundry/device-camera-go/commits/9d8893d))
- Remove Logging configuration ([#2f6e8f2](https://github.com/edgexfoundry/device-camera-go/commits/2f6e8f2))
- **security:** Get Camera creds from SecretProvider - Add implementation of getting camera credentials from secret provider / secret store instead of configuration toml directly - Remove plain text camera credentials from toml config file - Refactor camera credentials to allow multiple different camera credentials using DeviceList instead of Driver config ([#0fadece](https://github.com/edgexfoundry/device-camera-go/commits/0fadece))
### Style
- Use pascal case in the sample Profiles ([#b9d8927](https://github.com/edgexfoundry/device-camera-go/commits/b9d8927))
### Bug Fixes 🐛
- address Lenny's feedback about authMode none to skip getCredentials - skip GetCredentials from secret provider if AuthMethod is none ([#c058d96](https://github.com/edgexfoundry/device-camera-go/commits/c058d96))
- Remove retry items of SecretStore config and update secret path - go-mod-bootstrap has implemented the addition of prefix /v1/secret/edgex/ for the Path property of SecretStore config section, so we just use the service specific secret path in Toml files - removed retry related items in "SecretStore" config section ([#d51bbe5](https://github.com/edgexfoundry/device-camera-go/commits/d51bbe5))
- **build:** update go.mod to go 1.16 ([#d43abaa](https://github.com/edgexfoundry/device-camera-go/commits/d43abaa))
- **build:** update Dockerfiles to use go 1.16 ([#ba22a61](https://github.com/edgexfoundry/device-camera-go/commits/ba22a61))
- **snap:** '-go' suffix removed from device name ([#2a3a7a1](https://github.com/edgexfoundry/device-camera-go/commits/2a3a7a1))
- **snap:** run 'go mod tidy' ([#c9606bd](https://github.com/edgexfoundry/device-camera-go/commits/c9606bd))
- **snap:** update go to 1.16 ([#5bc5291](https://github.com/edgexfoundry/device-camera-go/commits/5bc5291))
- **snap:** update snap v2 support ([#5a3c97e](https://github.com/edgexfoundry/device-camera-go/commits/5a3c97e))
### Code Refactoring ♻
- bump dependency version and update import path ([#9562c59](https://github.com/edgexfoundry/device-camera-go/commits/9562c59))
- remove unimplemented InitCmd/RemoveCmd configuration ([#d1479f8](https://github.com/edgexfoundry/device-camera-go/commits/d1479f8))
- Remove the dependency of pkg/errors ([#1083b87](https://github.com/edgexfoundry/device-camera-go/commits/1083b87))
- Change PublishTopicPrefix value to be 'edgex/events/device' ([#0734506](https://github.com/edgexfoundry/device-camera-go/commits/0734506))
- use v2 SDK ([#3339319](https://github.com/edgexfoundry/device-camera-go/commits/3339319))
- Update to assign and uses new Port Assignments and service key ([#fe94c2d](https://github.com/edgexfoundry/device-camera-go/commits/fe94c2d))
    ```
    BREAKING CHANGE:
    Device Camera default port number has changed to 59985
    ```
    ```
    BREAKING CHANGE:
    Device Camera service key has changed to device-camera
    ```
### Documentation 📖
- update README ([#acb6ef3](https://github.com/edgexfoundry/device-camera-go/commits/acb6ef3))
- Add badges to readme ([#1610273](https://github.com/edgexfoundry/device-camera-go/commits/1610273))
### Build 👷
- update snap build ([#c46a067](https://github.com/edgexfoundry/device-camera-go/commits/c46a067))
- update build files for v2 ([#d808c09](https://github.com/edgexfoundry/device-camera-go/commits/d808c09))
- **snap:** update epoch and release version to Ireland ([#95a98a1](https://github.com/edgexfoundry/device-camera-go/commits/95a98a1))
### Continuous Integration 🔄
- update local docker image names ([#ac9394c](https://github.com/edgexfoundry/device-camera-go/commits/ac9394c))

<a name="v1.2.1"></a>
## [v1.2.1] - 2021-02-02
### Features ✨
- **snap:** add startup-duration and startup-interval configure options ([#089ba0f](https://github.com/edgexfoundry/device-camera-go/commits/089ba0f))
### Bug Fixes 🐛
- Set mediaType on all Binary resources ([#85](https://github.com/edgexfoundry/device-camera-go/issues/85)) ([#7c976a9](https://github.com/edgexfoundry/device-camera-go/commits/7c976a9))
### Continuous Integration 🔄
- add semantic.yml for commit linting, update PR template to latest ([#4530cb8](https://github.com/edgexfoundry/device-camera-go/commits/4530cb8))
- standardize dockerfiles ([#b2492c7](https://github.com/edgexfoundry/device-camera-go/commits/b2492c7))

<a name="v1.2.0"></a>
## [v1.2.0] - 2020-11-18
### Bug Fixes 🐛
- Changed to single qoutes per PR feedback ([#bfac45e](https://github.com/edgexfoundry/device-camera-go/commits/bfac45e))
- Correct logging setting for STDOUT ([#8d54db9](https://github.com/edgexfoundry/device-camera-go/commits/8d54db9))
- update incorrect edgexfoundry-holding imports ([#0e3ece9](https://github.com/edgexfoundry/device-camera-go/commits/0e3ece9))
- check for nil before adding an onvifClient ([#39108fe](https://github.com/edgexfoundry/device-camera-go/commits/39108fe))
- **snap:** Update snap versioning logic ([#4ae3ea6](https://github.com/edgexfoundry/device-camera-go/commits/4ae3ea6))
### Code Refactoring ♻
- Upgrade SDK to v1.2.4-dev.34 ([#832d354](https://github.com/edgexfoundry/device-camera-go/commits/832d354))
- update dockerfile to appropriately use ENTRYPOINT and CMD, closes[#43](https://github.com/edgexfoundry/device-camera-go/issues/43) ([#2df4ac3](https://github.com/edgexfoundry/device-camera-go/commits/2df4ac3))
### Build 👷
- Upgrade to Go1.15 ([#10b3fe6](https://github.com/edgexfoundry/device-camera-go/commits/10b3fe6))
- **all:** Enable use of DependaBot to maintain Go dependencies ([#2a98e4e](https://github.com/edgexfoundry/device-camera-go/commits/2a98e4e))

<a name="v1.1.2"></a>
## [v1.1.2] - 2020-08-19
### Features ✨
- **device-camera:** add snap package files for device-camera ([#a0ca43a](https://github.com/edgexfoundry/device-camera-go/commits/a0ca43a))
### Documentation 📖
- Add standart PR template ([#e63de38](https://github.com/edgexfoundry/device-camera-go/commits/e63de38))
