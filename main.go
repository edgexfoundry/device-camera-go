package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/edgexfoundry-holding/device-camera-go/cameradiscoveryprovider"
	"github.com/edgexfoundry/device-sdk-go"
	ds_models "github.com/edgexfoundry/device-sdk-go/pkg/models"
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
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

	// TODO: Populate all configuration from EdgeX registry?
	//  Currently we get application data for credentials from configuration.toml only.
	//  Otherwise assigned using environment vars in compose file, overridden by cmdline parameters.

	registryUrl = os.Getenv("dsdk_registry")
	confProfile = os.Getenv("dsdk_profile")
	confDir = os.Getenv("dsdk_confdir")

	// Process EdgeX DeviceService parameters
	flag.StringVar(&registryUrl, "registry", "", "Deprecated. Override consul registry url using Docker profile commandline parameter.")
	flag.StringVar(&registryUrl, "r", "", "Deprecated. Override consul registry url using Docker profile commandline parameter.")
	flag.StringVar(&confProfile, "profile", confProfile, "Specify a profile other than default. Examples are local and docker.")
	flag.StringVar(&confProfile, "p", confProfile, "Specify a profile other than default.")
	flag.StringVar(&confDir, "confdir", confDir, "Specify an alternate configuration directory.")
	flag.StringVar(&confDir, "c", confDir, "Specify an alternate configuration directory.")

	// Initial flag values are assigned by environment variables
	// This allows local host to set and also may be set using docker run (-e <variable=xxx>)
	// and using docker-compuse (add environment section within docker-compose.yml file).
	ip := os.Getenv("dcg_ip")
	netMask := os.Getenv("dcg_mask")
	scanDuration := os.Getenv("dcg_scanduration")
	intervalStr := os.Getenv("dcg_interval")
	source1 := os.Getenv("dcg_source1")
	tagsFile := os.Getenv("dcg_tagsFile")
	cameraCredentialsFile := os.Getenv("dcg_cameracredentials")

	// Assign default values to be used for certain device-camera-go parameters 
	// if not specified by commandline / environment
	if len(ip) == 0 {
		ip = "192.168.0.1-30"
	}
	if len(scanDuration) == 0 {
		scanDuration = "15s"
	}
	// Convert env string to int
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 60
	}
	if len(tagsFile) == 0 {
		tagsFile = "./res/tags.json"
	}
	if len(cameraCredentialsFile) == 0 {
		cameraCredentialsFile = "./res/cam-credentials.conf"
	}
	if len(source1) != 0 {
		// assign a single source (with n ports) using environment
		// extend as you deem useful by adding environment vars, etc.
		sourceFlags = append(sourceFlags, source1)
	} else {
		flag.Var(&sourceFlags, "source", "source to scan, -source axis:554 -source onvif:80")
	}

	// Process remaining CameraDiscovery parameters
	flag.StringVar(&ip, "ip", ip, "IP address for nmap scan")
	flag.StringVar(&netMask, "mask", netMask, "Network mask for nmap scan")
	flag.StringVar(&scanDuration, "scanduration", scanDuration, "Duration to permit for each network discovery scan; -scanduration \"10s\"")
	flag.IntVar(&interval, "interval", interval, "Interval between discovery scans, in seconds. Must be > scanduration; -interval 180")
	flag.StringVar(&tagsFile, "tagsFile", tagsFile, "Location of file where camera tags are cached")
	flag.StringVar(&cameraCredentialsFile, "cameracredentials", cameraCredentialsFile,
		"Path to file containing credentials for the camera (user and password, tab-separated)")
	flag.Parse()

	camInfoFileOnvif := "./res/caminfocache.json"
	camInfoFileAxis := "./res/caminfocache-axis.json"

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
		CredentialsFile:  cameraCredentialsFile, // used for explicit (e.g., in case separate file desired)
		ConfDir:          confDir, // used for credentials default/fallback
		ConfProfile:      confProfile, // used for credentials default/fallback
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
	fmt.Println(fmt.Sprintf("IPRange: %v", options.IP))
	fmt.Println(fmt.Sprintf("NetMask: %v", options.NetMask))
	fmt.Println(fmt.Sprintf("Interval: %v", options.Interval))
	fmt.Println(fmt.Sprintf("ScanDuration: %v", options.ScanDuration))
	fmt.Println(fmt.Sprintf("CredFile: %v", options.Credentials))

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


