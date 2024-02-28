package fabconnect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteFabconnectConfig(t *testing.T) {
	directory := t.TempDir()

	testCases := []struct {
		Name     string
		filePath string
	}{
		{Name: "TestPath-1", filePath: directory + "/fabconfig1.yaml"},
		{Name: "TestPath-2", filePath: directory + "/fabconfig2.yaml"},
		{Name: "TestPath-3", filePath: directory + "/fabconfig3.yaml"},
		{Name: "TestPath-4", filePath: directory + "/fabconfig4.yaml"},
	}
	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			err := WriteFabconnectConfig(tc.filePath)
			if err != nil {
				t.Log("cannot write config:", err)
			}
		})
		assert.NotNil(t, tc.filePath)
	}
}
