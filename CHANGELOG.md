
<a name="EdgeX Camera Device Service (found in device-camera-go) Changelog"></a>
## EdgeX Camera Device Service
[Github repository](https://github.com/edgexfoundry/device-camera-go)

### Change Logs for EdgeX Dependencies
- [device-sdk-go](https://github.com/edgexfoundry/device-sdk-go/blob/main/CHANGELOG.md)
- [go-mod-core-contracts](https://github.com/edgexfoundry/go-mod-core-contracts/blob/main/CHANGELOG.md)
- [go-mod-bootstrap](https://github.com/edgexfoundry/go-mod-bootstrap/blob/main/CHANGELOG.md)
- [go-mod-messaging](https://github.com/edgexfoundry/go-mod-messaging/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-registry](https://github.com/edgexfoundry/go-mod-registry/blob/main/CHANGELOG.md)  (indirect dependency)
- [go-mod-secrets](https://github.com/edgexfoundry/go-mod-secrets/blob/main/CHANGELOG.md) (indirect dependency)
- [go-mod-configuration](https://github.com/edgexfoundry/go-mod-configuration/blob/main/CHANGELOG.md) (indirect dependency)

## [v2.2.0] Kamakura - 2022-05-11  (Not Compatible with 1.x releases)

### Features ✨
- Enable security hardening ([#583c54f](https://github.com/edgexfoundry/device-camera-go/commits/583c54f))

### Bug Fixes 🐛
- **snap:** Expose parent directory in device-config plug ([#f6f2141](https://github.com/edgexfoundry/device-camera-go/commits/f6f2141))

### Code Refactoring ♻
- **snap:** Remove redundant content indentifier ([#a81407d](https://github.com/edgexfoundry/device-camera-go/commits/a81407d))

### Documentation 📖
- **snap:** remove two environment configuration overrides ([#ccfe6b2](https://github.com/edgexfoundry/device-camera-go/commits/ccfe6b2))

### Build 👷
- Update to latest SDK w/o ZMQ on windows ([#9b3fa78](https://github.com/edgexfoundry/device-camera-go/commits/9b3fa78))
    ```
    BREAKING CHANGE:
    ZeroMQ no longer supported on native Windows for EdgeX
    MessageBus
    ```
- **snap:** source metadata from central repo ([#193](https://github.com/edgexfoundry/device-camera-go/issues/193)) ([#3f3ab22](https://github.com/edgexfoundry/device-camera-go/commits/3f3ab22))

### Continuous Integration 🔄
- gomod changes related for Go 1.17 ([#82a10a5](https://github.com/edgexfoundry/device-camera-go/commits/82a10a5))
- Go 1.17 related changes ([#569337f](https://github.com/edgexfoundry/device-camera-go/commits/569337f))

## [v2.1.0] Jakarta - 2021-11-18  (Not Compatible with 1.x releases)

### Features ✨
- Update config files to include secretsfile ([#94ad87a](https://github.com/edgexfoundry/device-camera-go/commits/94ad87a))

### Documentation 📖
- Add comment for secretsfile ([#6f6b908](https://github.com/edgexfoundry/device-camera-go/commits/6f6b908))

### Build 👷
- Update to released SDK and go-mods ([#4ddf907](https://github.com/edgexfoundry/device-camera-go/commits/4ddf907))
- Bump to device-sdk-go[@v2](:/v2).0.1-dev.23 and go-mod-core-contracts[@v2](:/v2).0.1-dev.26 ([#89ba74a](https://github.com/edgexfoundry/device-camera-go/commits/89ba74a))
- **snap:** upgrade base to core20 ([#ee1913c](https://github.com/edgexfoundry/device-camera-go/commits/ee1913c))
- **snap:** bump edgex-snap-hooks to support secretsfile config ([#e3dffcc](https://github.com/edgexfoundry/device-camera-go/commits/e3dffcc))

## [v2.0.1] Ireland - 2021-10-08  (Not Compatible with 1.x releases)
- Use correct constants for authentication ([#4e5edb8](https://github.com/edgexfoundry/device-camera-go/commits/4e5edb8))
- Update PascalCase for device resource names ([#138](https://github.com/edgexfoundry/device-camera-go/issues/138)) - Aligns code with device profile changes made in PR 111 ([#8e9dd4c](https://github.com/edgexfoundry/device-camera-go/commits/8e9dd4c))
- Update all TOML to use quote and not single-quote ([#d3b9f4b](https://github.com/edgexfoundry/device-camera-go/commits/d3b9f4b))

## [v2.0.0] Ireland - 2021-06-30  (Not Compatible with 1.x releases)

### Features ✨
- Enable using MessageBus as the default ([#9d8893d](https://github.com/edgexfoundry/device-camera-go/commits/9d8893d))
- Remove Logging configuration ([#2f6e8f2](https://github.com/edgexfoundry/device-camera-go/commits/2f6e8f2))
- **security:** Get Camera creds from SecretProvider - Add implementation of getting camera credentials from secret provider / secret store instead of configuration toml directly - Remove plain text camera credentials from toml config file - Refactor camera credentials to allow multiple different camera credentials using DeviceList instead of Driver config ([#0fadece](https://github.com/edgexfoundry/device-camera-go/commits/0fadece))

### Style
- Use pascal case in the sample Profiles ([#b9d8927](https://github.com/edgexfoundry/device-camera-go/commits/b9d8927))

### Code Refactoring ♻
- remove unimplemented InitCmd/RemoveCmd configuration ([#d1479f8](https://github.com/edgexfoundry/device-camera-go/commits/d1479f8))
- Remove the dependency of pkg/errors ([#1083b87](https://github.com/edgexfoundry/device-camera-go/commits/1083b87))
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
- update go.mod to go 1.16 ([#d43abaa](https://github.com/edgexfoundry/device-camera-go/commits/d43abaa))
- update Dockerfiles to use go 1.16 ([#ba22a61](https://github.com/edgexfoundry/device-camera-go/commits/ba22a61))
- **snap:** '-go' suffix removed from device name ([#2a3a7a1](https://github.com/edgexfoundry/device-camera-go/commits/2a3a7a1))
- **snap:** run 'go mod tidy' ([#c9606bd](https://github.com/edgexfoundry/device-camera-go/commits/c9606bd))
- **snap:** update go to 1.16 ([#5bc5291](https://github.com/edgexfoundry/device-camera-go/commits/5bc5291))
- **snap:** update snap v2 support ([#5a3c97e](https://github.com/edgexfoundry/device-camera-go/commits/5a3c97e))
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
