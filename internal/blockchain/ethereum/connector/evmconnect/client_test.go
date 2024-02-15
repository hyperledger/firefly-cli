package evmconnect

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPort(t *testing.T) {
	t.Run("testPort", func(t *testing.T) {
		Port := 5008
		e := &Evmconnect{}
		PortInt := e.Port()
		assert.Equal(t, Port, PortInt)
	})
}

func TestName(t *testing.T) {
	t.Run("testName", func(t *testing.T) {
		Name := "evmconnect"
		e := &Evmconnect{}
		NameStr := e.Name()
		assert.Equal(t, Name, NameStr)
	})
}

func TestNewEvmconnect(t *testing.T) {
	var Ctx context.Context
	EvmConnect := NewEvmconnect(Ctx)
	assert.NotNil(t, EvmConnect)

}
