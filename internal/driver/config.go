package driver

import (
	"flag"
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
)

type configuration struct {
	Camera cameraInfo
}

type cameraInfo struct {
	User       string
	Password   string
	AuthMethod string
}

// loadConfigFromFile use to load toml configuration
func loadConfigFromFile() (*configuration, error) {
	config := new(configuration)

	confDir := flag.Lookup("confdir").Value.(flag.Getter).Get().(string)
	if len(confDir) == 0 {
		confDir = flag.Lookup("c").Value.(flag.Getter).Get().(string)
	}

	if len(confDir) == 0 {
		confDir = "./res"
	}

	filePath := fmt.Sprintf("%v/configuration-driver.toml", confDir)

	file, err := ioutil.ReadFile(filePath)
	if err != nil {
		return config, fmt.Errorf("could not load configuration file (%s): %v", filePath, err.Error())
	}

	err = toml.Unmarshal(file, config)
	if err != nil {
		return config, fmt.Errorf("unable to parse configuration file (%s): %v", filePath, err.Error())
	}
	return config, err
}
