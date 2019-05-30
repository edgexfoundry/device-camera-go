package main

import (
	"flag"
	"fmt"
	"github.com/edgexfoundry/device-sdk-go"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
	"github.com/edgexfoundry-holding/device-camera-go/cameradiscoveryprovider"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const (
	version     string = "0.1"
	serviceName string = "device-camera-go"
)

var (
	confProfile string
	confDir     string
	registryUrl string
)

// SourceList represents the parameters instructing to scan for a particular
// camera source interface; e.g., onvif, axis,
type SourceList []string

var sourceFlags SourceList

func (list *SourceList) String() string {
	return ""
}

// Set is overriding SourceList interface method
func (list *SourceList) Set(val string) error {
	*list = append(*list, val)
	return nil
}

func main() {
	// Default device-camera-go parameters

	// IP Range default
	ip := "192.168.0.1-30"
	netMask := ""
	scanDuration := "15s"
	interval := 60

	camInfoFileOnvif := "./res/caminfocache.json"
	camInfoFileAxis := "./res/caminfocache-axis.json"
	tagsFile := "./res/tags.json"
	cameraCredentialsFile := "./res/cam-credentials.conf"
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// Process EdgeX DeviceService parameters
	flag.StringVar(&registryUrl, "registry", "", "Deprecated. Override consul registry url using Docker profile commandline parameter.")
	flag.StringVar(&registryUrl, "r", "", "Deprecated. Override consul registry url using Docker profile commandline parameter.")
	flag.StringVar(&confProfile, "profile", "", "Specify a profile other than default.")
	flag.StringVar(&confProfile, "p", "", "Specify a profile other than default.")
	flag.StringVar(&confDir, "confdir", "", "Specify an alternate configuration directory.")
	flag.StringVar(&confDir, "c", "", "Specify an alternate configuration directory.")

	// Process CameraDiscovery parameters
	flag.StringVar(&ip, "ip", ip, "IP address for nmap scan")
	flag.StringVar(&netMask, "mask", netMask, "Network mask for nmap scan")
	flag.StringVar(&scanDuration, "scanduration", scanDuration, "Duration to permit for each network discovery scan; -scanduration \"10s\"")
	flag.IntVar(&interval, "interval", interval, "Interval between discovery scans, in seconds. Must be > scanduration; -interval 180")
	flag.Var(&sourceFlags, "source", "source to scan, -source axis:554 -source onvif:80")
	flag.StringVar(&tagsFile, "tagsFile", tagsFile, "Location of file where camera tags are cached")
	flag.StringVar(&cameraCredentialsFile, "cameracredentials", cameraCredentialsFile,
		"Path to file containing credentials for the camera (user and password, tab-separated)")
	flag.Parse()

	camInfoCache := cameradiscoveryprovider.CamInfo{}
	tagCache := cameradiscoveryprovider.Tags{}

	var supportedSources []cameradiscoveryprovider.CameraSource
	supportedSources = []cameradiscoveryprovider.CameraSource{
		{
			Index:            cameradiscoveryprovider.ONVIF,
			Name:             "onvif",
			DeviceNamePrefix: "edgex-camera-onvif-",
			ProfileName:      "camera-profile-onvif",
			DefaultPort:      80,
		},
		{
			Index:            cameradiscoveryprovider.Axis,
			Name:             "axis",
			DeviceNamePrefix: "edgex-camera-axis-",
			ProfileName:      "camera-profile-axis",
			DefaultPort:      554,
		},
	}
	options := cameradiscoveryprovider.Options{
		Interval:         interval,
		ScanDuration:     scanDuration,
		SourceFlags:      sourceFlags,
		IP:               ip,
		NetMask:          netMask,
		Credentials:      cameraCredentialsFile,
		SupportedSources: supportedSources,
	}
	ac := cameradiscoveryprovider.AppCache{
		CamInfoCache:  &camInfoCache,
		InfoFileOnvif: camInfoFileOnvif,
		InfoFileAxis:  camInfoFileAxis,
		TagCache:      &tagCache,
		TagsFile:      tagsFile,
	}

	sd := cameradiscoveryprovider.CameraDiscoveryProvider{}
	// NOTE: If we were only interested to load device configurations at startup using .YML resources, we could use:
	// startup.Bootstrap(serviceName, version, &sd)

	if err := startService(serviceName, version, &sd, &options, &ac); err != nil {
		Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// Fprintf method overrides fmt.Fprintf to perform error check in one place
//func Fprintf(w io.Writer, format string, a ...interface{}) (n int, err error) {
func Fprintf(w io.Writer, format string, a ...interface{}) {
	_, err := fmt.Fprintf(w, format, a...)
	if err != nil {
		log.Fatalf("Fatal error accessing output target: %s", err)
	}
}

func startService(serviceName string, serviceVersion string, driver ds_models.ProtocolDriver, options *cameradiscoveryprovider.Options, ac *cameradiscoveryprovider.AppCache) error {
	// Instantiate our CameraDiscovery provider "device"
	cameradiscoveryprovider := cameradiscoveryprovider.New(options, ac)
	fmt.Println(fmt.Sprintf("SourceFlags: %v", options.SourceFlags))
	// Create EdgeX service and pass in the provider "device"
	deviceService, err := device.NewService(serviceName, serviceVersion, confProfile, confDir, registryUrl, cameradiscoveryprovider)
	if err != nil {
		return err
	}
	errChan := make(chan error, 2)
	listenForInterrupt(errChan)
	// Start the EdgeX service
	Fprintf(os.Stdout, "Calling service.Start.\n")
	err = deviceService.Start(errChan)
	if err != nil {
		return err
	}
	err = <-errChan
	Fprintf(os.Stdout, "Terminating: %v.\n", err)
	return deviceService.Stop(false)
}

func listenForInterrupt(errChan chan error) {
	go func() {
		// Set up channel on which to send signal notifications.
		// We must use a buffered channel or risk missing the signal
		// if we're not ready to receive when the signal is sent.
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errChan <- fmt.Errorf("%s", <-c)
	}()
}
