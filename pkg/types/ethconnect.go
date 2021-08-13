package types

type Contract struct {
	ContractName string      `json:"contractName"`
	ABI          interface{} `json:"abi"`
	Bytecode     string      `json:"bytecode"`
}
