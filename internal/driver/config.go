/*******************************************************************************
 * Copyright 2021 Intel Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *
 *******************************************************************************/

package driver

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/v2/models"
)

type configuration struct {
	CredentialsRetryTime int
	CredentialsRetryWait int
}

type cameraInfo struct {
	Address         string
	AuthMethod      string
	CredentialPaths string
}

// loadCameraConfig loads the camera configuration
func loadCameraConfig(configMap map[string]string) (*configuration, error) {
	config := new(configuration)
	err := load(configMap, config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func CreateCameraInfo(protocols map[string]models.ProtocolProperties) (*cameraInfo, error) {
	info := new(cameraInfo)
	protocol, ok := protocols[HTTP_PROTOCOL]
	if !ok {
		return info, errors.New("unable to load config, Protocol HTTP not exist")
	}

	if err := load(protocol, info); err != nil {
		return info, err
	}

	return info, nil
}

// load by reflect to check map key and then fetch the value
func load(config map[string]string, des interface{}) error {
	errorMessage := "unable to load config, '%s' not exist"
	val := reflect.ValueOf(des).Elem()
	for i := 0; i < val.NumField(); i++ {
		typeField := val.Type().Field(i)
		valueField := val.Field(i)

		val, ok := config[typeField.Name]
		if !ok {
			return fmt.Errorf(errorMessage, typeField.Name)
		}

		switch valueField.Kind() {
		case reflect.Int:
			intVal, err := strconv.Atoi(val)
			if err != nil {
				return err
			}
			valueField.SetInt(int64(intVal))
		case reflect.String:
			valueField.SetString(val)
		default:
			return fmt.Errorf("none supported value type %v ,%v", valueField.Kind(), typeField.Name)
		}
	}
	return nil
}
