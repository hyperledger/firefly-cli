// Copyright © 2024 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sync"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/sirupsen/logrus"
)

type TestHelper struct {
	FabricURL     string
	EthConnectURL string
	EvmConnectURL string
}

var logMutex sync.Mutex

var (
	FabricEndpoint     = "http://localhost:7054"
	EthConnectEndpoint = "http://localhost:8080"
	EvmConnectEndpoint = "http://localhost:5008"
)

func StartMockServer(t *testing.T) {
	httpmock.Activate()
}

// mockprotocol endpoints for testing
func NewTestEndPoint(t *testing.T) *TestHelper {
	return &TestHelper{
		FabricURL:     FabricEndpoint,
		EthConnectURL: EthConnectEndpoint,
		EvmConnectURL: EvmConnectEndpoint,
	}
}

func StopMockServer(_ *testing.T) {
	httpmock.DeactivateAndReset()
}

// checks if exp value and act value are  equal
func Equals(tb testing.TB, exp, act interface{}) {
	if !reflect.DeepEqual(exp, act) {
		_, file, line, _ := runtime.Caller(1)
		fmt.Printf("\033[31m%s:%d:\n\n\texp: %#v\n\n\tgot: %#v\033[39m\n\n", filepath.Base(file), line, exp, act)
		tb.FailNow()
	}
}

// ReadFileToString reads the contents of a file and returns it as a string.
func ReadFileToString(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// CaptureOutput redirects the standard output to a buffer and returns the original output writer and the captured output.
func CaptureOutput() (*os.File, *bytes.Buffer) {
	originalOutput := os.Stdout // Save the original output

	// Create a pipe to capture the output
	_, writer, _ := os.Pipe()
	os.Stdout = writer

	// Create a buffer to capture the output
	buffer := &bytes.Buffer{}

	// Redirect logrus output to the same buffer
	logMutex.Lock()
	logrus.SetOutput(io.MultiWriter(originalOutput, buffer))
	logMutex.Unlock()

	return originalOutput, buffer
}
