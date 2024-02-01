package ethereum

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTruffleCompiledContract(t *testing.T) {
	dir := "testdata"
	contracFile := filepath.Join(dir, "truffle.json")

	t.Run("TestTruffleCompilesContract", func(t *testing.T) {
		ExpectedContractName := "FireFly_Client"

		compiledContracts, err := ReadTruffleCompiledContract(contracFile)
		if err != nil {
			t.Logf("unable to read truffle contract : %v", err)
		}
		contractMap := compiledContracts.Contracts
		assert.NotNil(t, compiledContracts)
		assert.NotNil(t, contractMap)

		contractName, ok := contractMap[ExpectedContractName]
		assert.True(t, ok, "Expected contract '%s' not found", ExpectedContractName)
		assert.NotNil(t, contractName)
	})
}

func TestReadSolCompiledContract(t *testing.T){
	
}