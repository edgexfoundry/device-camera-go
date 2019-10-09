package driver

import (
	"errors"
	"reflect"
	"testing"

	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
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

type mockLogger struct {
	debug, error, info, trace, warn int
}

func (m *mockLogger) SetLogLevel(logLevel string) error {
	panic("implement me")
}

func (m *mockLogger) Debug(msg string, args ...interface{}) {
	m.debug--
}

func (m *mockLogger) Error(msg string, args ...interface{}) {
	m.error--
}

func (m *mockLogger) Info(msg string, args ...interface{}) {
	m.info--
}

func (m *mockLogger) Trace(msg string, args ...interface{}) {
	m.trace--
}

func (m *mockLogger) Warn(msg string, args ...interface{}) {
	m.warn--
}

func (m mockLogger) VerifyLoggerCalls() bool {
	return m.debug == 0 && m.error == 0 && m.info == 0 && m.trace == 0 && m.warn == 0
}

func TestDriver_addrFromProtocols(t *testing.T) {
	tests := []struct {
		name          string
		protocols     map[string]contract.ProtocolProperties
		logger        *mockLogger
		expectedValue string
		expectedError bool
	}{
		{
			name:          "OK",
			protocols:     map[string]contract.ProtocolProperties{"HTTP": {"Address": "someaddress"}},
			logger:        &mockLogger{},
			expectedValue: "someaddress",
			expectedError: false,
		},
		{
			name:          "Missing HTTP protocol",
			protocols:     map[string]contract.ProtocolProperties{"TCP": {"Address": "address2"}},
			logger:        &mockLogger{error: 1},
			expectedValue: "",
			expectedError: true,
		},
		{
			name:          "Missing address",
			protocols:     map[string]contract.ProtocolProperties{"HTTP": {"Secure": "True"}},
			logger:        &mockLogger{error: 1},
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

			if !test.logger.VerifyLoggerCalls() {
				t.Errorf("Didn't receive expected writes to logger")
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
