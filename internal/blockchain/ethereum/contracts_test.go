package ethereum

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadTruffleCompiledContract(t *testing.T) {
	dir := "testdata"
	contractFile := filepath.Join(dir, "truffle.json")
	t.Run("TestTruffleCompilesContract", func(t *testing.T) {
		ExpectedContractName := "FireFly_Client"

		compiledContracts, err := ReadTruffleCompiledContract(contractFile)
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

func TestReadSolCompiledContract(t *testing.T) {
	dir := "testdata"
	contractFile := filepath.Join(dir, "sol.json")
	t.Run("TestReadSolContract", func(t *testing.T) {
		SolContract, err := ReadSolcCompiledContract(contractFile)
		if err != nil {
			t.Logf("Unable to read sol contract: %v", err)
		}
		assert.NotNil(t, SolContract)
		contractMap := SolContract.Contracts
		assert.NotNil(t, contractMap)
	})
}
