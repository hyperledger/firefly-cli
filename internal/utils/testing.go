// This file contains different Setup and tools, for the FireFly-CLI testing Environment
package utils

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"

	"github.com/jarcoal/httpmock"
)

type TestHelper struct {
	FabricURL     string
	EthConnectURL string
	EvmConnectURL string
}

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

	return originalOutput, buffer
}