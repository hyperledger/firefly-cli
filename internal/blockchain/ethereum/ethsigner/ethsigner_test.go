package ethsigner

import (
	"testing"

	"github.com/hyperledger/firefly-cli/pkg/types"
)

func TestWriteSignerConfig(t *testing.T) {
	options := &types.InitOptions{ChainID: int64(689)}
	stack := &types.Stack{Name: "firefly_eth"}
	rpcURL := "http://localhost:9583"
	e := EthSignerProvider{
		stack: stack,
	}
	err := e.WriteConfig(options, rpcURL)
	if err != nil {
		t.Logf("unable to write config :%v", err)
	}
}
