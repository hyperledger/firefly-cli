package ethconnect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPort(t *testing.T) {
	t.Run("testPort", func(t *testing.T) {
		Port := 8080
		E := &Ethconnect{}
		PortInt := E.Port()
		assert.Equal(t, Port, PortInt)
	})
}

func TestName(t *testing.T) {
	t.Run("testName", func(t *testing.T) {
		Name := "ethconnect"
		E := &Ethconnect{}
		NameStr := E.Name()
		assert.Equal(t, Name, NameStr)
	})
}
