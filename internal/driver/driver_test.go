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
	"reflect"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

type testStringer struct {
	str string
	err error
}

func (ts testStringer) StringValue() (string, error) {
	if ts.err != nil {
		return "", ts.err
	}

	return ts.str, nil
}

func newTestStringer(str string, err error) testStringer {
	ts := testStringer{
		str: str,
		err: err,
	}

	return ts
}

func TestDriver_addrFromProtocols(t *testing.T) {
	tests := []struct {
		name          string
		protocols     map[string]models.ProtocolProperties
		logger        logger.LoggingClient
		expectedValue string
		expectedError bool
	}{
		{
			name: "OK",
			protocols: map[string]models.ProtocolProperties{HTTP_PROTOCOL: {
				"Address":         "someaddress",
				"AuthMethod":      "usernamepassword",
				"CredentialsPath": "secrets"}},
			logger:        logger.NewMockClient(),
			expectedValue: "someaddress",
			expectedError: false,
		},
		{
			name:          "Missing HTTP protocol",
			protocols:     map[string]models.ProtocolProperties{"TCP": {"Address": "address2"}},
			logger:        logger.NewMockClient(),
			expectedValue: "",
			expectedError: true,
		},
		{
			name:          "Missing address",
			protocols:     map[string]models.ProtocolProperties{HTTP_PROTOCOL: {"Secure": "True"}},
			logger:        logger.NewMockClient(),
			expectedValue: "",
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := Driver{lc: test.logger}
			res, err := driver.addrFromProtocols(test.protocols)

			if err != nil && !test.expectedError {
				t.Errorf("Unexpected error: %v", err.Error())
				return
			} else if err == nil && test.expectedError {
				t.Errorf("Expected an error but didn't get one")
				return
			}

			if res != test.expectedValue {
				t.Errorf("Expected '%v', Received '%v'", test.expectedValue, res)
			}
		})
	}
}

func TestIn(t *testing.T) {
	tests := []struct {
		name      string
		needle    string
		haystack  []string
		expResult bool
	}{
		{
			name:      "single element equals",
			needle:    "bosch",
			haystack:  []string{"bosch"},
			expResult: true,
		},
		{
			name:      "multi-element equals",
			needle:    "bosch",
			haystack:  []string{"axis", "bosch"},
			expResult: true,
		},
		{
			name:      "empty list",
			needle:    "bosch",
			haystack:  []string{},
			expResult: false,
		},
		{
			name:      "not found",
			needle:    "bosch",
			haystack:  []string{"biggio", "bagwell", "bell"},
			expResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res := in(test.needle, test.haystack)
			if res != test.expResult {
				t.Errorf("Expected: '%v', Result: '%v'", test.expResult, res)
			}
		})
	}
}

func TestStructFromParam(t *testing.T) {
	type user struct {
		Nickname    string
		Password    string
		IsAdmin     bool
		Permissions []string
	}
	users := []user{
		{}, // zero value
		{
			Nickname: "johndoe",
			Password: "abc123",
			IsAdmin:  true,
		},
		{
			Nickname:    "janedoe",
			Password:    "secure",
			IsAdmin:     false,
			Permissions: []string{"foo", "bar", "baz"},
		},
	}
	tests := []struct {
		name           string
		stringer       testStringer
		dest           user
		expectedResult user
		expectedError  bool
	}{
		{
			name:           "simple struct",
			stringer:       newTestStringer("{\"Nickname\": \"johndoe\", \n\"Password\":\"abc123\",\n\"IsAdmin\":true}", nil),
			dest:           users[0],
			expectedResult: users[1],
			expectedError:  false,
		},
		{
			name:           "struct with list",
			stringer:       newTestStringer("{\"Nickname\": \"janedoe\", \n\"Password\":\"secure\",\n\"IsAdmin\":false,\n\"Permissions\":[\"foo\",\"bar\",\"baz\"]}", nil),
			dest:           users[0],
			expectedResult: users[2],
			expectedError:  false,
		},
		{
			name:           "expected error from unmarshal",
			stringer:       newTestStringer("ipsum lorum", nil),
			dest:           users[0],
			expectedResult: users[0],
			expectedError:  true,
		},
		{
			name:           "error from stringer",
			stringer:       newTestStringer("", errors.New("some error")),
			dest:           users[0],
			expectedResult: users[0],
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := structFromParam(test.stringer, &test.dest)
			if (err != nil) != test.expectedError {
				t.Errorf("Unexpected error: %v", err)
				t.Fail()
				return
			}

			if !reflect.DeepEqual(test.dest, test.expectedResult) {
				t.Errorf("Unmarshaled struct (%v) doesn't match expectation.", test.dest)
				t.Fail()
				return
			}
		})
	}
}
