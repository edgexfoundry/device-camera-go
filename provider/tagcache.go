// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2018-2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

// Package cameradiscoveryprovider implements a CameraDiscovery provider which scans
// network for existence of compatible/requested IP cameras and registers these
// as EdgeX devices, with corresponding management/control interfaces.
package cameradiscoveryprovider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
)

// Tags is a struct containing tags to be assigned to cameras in a form of
// map where serial number is the key and tags map is the value
type Tags struct {
	Tags map[string]map[string]interface{}
}

// LoadTags loads tags from a json file
func (st *Tags) LoadTags(filePath string) error {
	bytes, err := ioutil.ReadFile(filepath.Clean(filePath))
	if err != nil {
		fmt.Println("Loading file failed. Using empty tags")
		st.Tags = map[string]map[string]interface{}{}
		return err
	}
	err = json.Unmarshal(bytes, &(st.Tags))
	if err != nil {
		fmt.Println("Parsing JSON failed. Using empty tags")
		st.Tags = map[string]map[string]interface{}{}
	}
	return err
}

// SaveTags saves tags as a json file
func (st *Tags) SaveTags(filePath string) error {
	tagsJSON, err := json.Marshal(&st.Tags)
	if err != nil {
		fmt.Println("Saving tags as JSON failed. Using runtime tag cache only")
		return err
	}
	err = ioutil.WriteFile(filePath, tagsJSON, 0600)
	if err != nil {
		fmt.Println("Saving file failed. Using runtime tag cache only")
		return err
	}
	return nil
}
