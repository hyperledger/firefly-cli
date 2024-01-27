package tezosconnect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPort(t *testing.T) {
	t.Run("testPort", func(t *testing.T) {
		Port := 5008
		T := &Tezosconnect{}
		PortInt := T.Port()
		assert.Equal(t, Port, PortInt)
	})
}

func TestName(t *testing.T) {
	t.Run("testName", func(t *testing.T) {
		Name := "tezosconnect"
		T := &Tezosconnect{}
		NameStr := T.Name()
		assert.Equal(t, Name, NameStr)
	})
}
