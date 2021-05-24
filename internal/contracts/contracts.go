package contracts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Contract struct {
	ContractName string      `json:"contractName"`
	ABI          interface{} `json:"abi"`
	Bytecode     string      `json:"bytecode"`
}

type ethconnectRequestHeaders struct {
	Type string `json:"type,omitempty"`
}

type ethconnectRequestBody struct {
	Headers      *ethconnectRequestHeaders `json:"headers,omitempty"`
	From         string                    `json:"from,omitempty"`
	Params       []string                  `json:"params,omitempty"`
	Gas          int                       `json:"gas,omitempty"`
	ContractName string                    `json:"contractName,omitempty"`
	ABI          interface{}               `json:"abi,omitempty"`
	Compiled     string                    `json:"compiled"`
}

type PublishAbiResponseBody struct {
	ID string `json:"id,omitempty"`
}

/*
{
    "created": "2021-05-21T16:09:47Z",
    "id": "70710a68-b9c5-4600-6104-d944b61f82fd",
    "name": "",
    "description": "",
    "path": "/abis/70710a68-b9c5-4600-6104-d944b61f82fd",
    "deployable": true,
    "openapi": "http://127.0.0.1:8080/abis/70710a68-b9c5-4600-6104-d944b61f82fd?swagger",
    "compilerVersion": ""
}
*/

type DeployContractResponseBody struct {
	ContractAddress string `json:"contractAddress,omitempty"`
}

/*
{
    "headers": {
        "id": "e4926411-2179-4c40-63c2-3e3aba8ac905",
        "type": "TransactionSuccess",
        "timeReceived": "2021-05-21T16:37:58.707293Z",
        "timeElapsed": 0.917213,
        "requestOffset": "",
        "requestId": "70710a68-b9c5-4600-6104-d944b61f82fd"
    },
    "blockHash": "0xafff7466fa527176799020cabeb06368f478a0a687cca0b38ff64cff88b0e423",
    "blockNumber": "1",
    "openapi": "http://127.0.0.1:8080/contracts/af1292f453d7f33f0c2c217968879deb435cbf3a?openapi",
    "apiexerciser": "http://127.0.0.1:8080/contracts/af1292f453d7f33f0c2c217968879deb435cbf3a?ui",
    "contractAddress": "0xaf1292f453d7f33f0c2c217968879deb435cbf3a",
    "cumulativeGasUsed": "1106739",
    "from": "0x58b49fc734fc3f90a5db18456d656ca104c8d351",
    "gasUsed": "1106739",
    "nonce": "0",
    "status": "1",
    "to": null,
    "transactionHash": "0x092c2573b6c5278d17b44964b178985e0c2e8fbd0ba5e1c175894bdebad1b163",
    "transactionIndex": "0"
}
*/

type RegisterResponseBody struct {
	Created      string `json:"created,omitempty"`
	Address      string `json:"string,omitempty"`
	Path         string `json:"path,omitempty"`
	ABI          string `json:"ABI,omitempty"`
	OpenAPI      string `json:"openapi,omitempty"`
	RegisteredAs string `json:"registeredAs,omitempty"`
}

/*
{
    "created": "2021-05-21T17:16:05Z",
    "address": "af1292f453d7f33f0c2c217968879deb435cbf3a",
    "path": "/contracts/banana",
    "abi": "70710a68-b9c5-4600-6104-d944b61f82fd",
    "openapi": "http://127.0.0.1:8080/contracts/banana?swagger",
    "registeredAs": "banana"
}
*/

func ReadCompiledContract(filePath string) *Contract {
	d, _ := ioutil.ReadFile(filePath)
	var contract *Contract
	err := json.Unmarshal(d, &contract)
	if err != nil {
		fmt.Println(err.Error())
	}
	return contract
}

// func DeployToEthconnect(contract *Contract, ethconnectUrl string, walletAddress string, params []string) *DeployContractResponseBody {

// 	byteCodeHex, _ := hex.DecodeString(contract.Bytecode[2:])
// 	byteCodeBase64 := base64.StdEncoding.EncodeToString(byteCodeHex)

// 	requestBody := &ethconnectRequestBody{
// 		Headers: &ethconnectRequestHeaders{
// 			Type: "DeployContract",
// 		},
// 		ContractName: contract.ContractName,
// 		From:         walletAddress,
// 		Params:       params,
// 		ABI:          contract.ABI,
// 		Compiled:     byteCodeBase64,
// 	}
// 	requestBodyBytes, _ := json.Marshal(requestBody)

// 	req, _ := http.NewRequest("POST", ethconnectUrl, bytes.NewBuffer(requestBodyBytes))
// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("x-kaleido-register", "foo")
// 	req.Header.Set("x-kaleido-sync", "true")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		panic(err)
// 	}
// 	defer resp.Body.Close()

// 	fmt.Println("response Status:", resp.Status)
// 	fmt.Println("response Headers:", resp.Header)
// 	body, _ := ioutil.ReadAll(resp.Body)
// 	fmt.Println(string(body))
// 	return nil
// 	// var responseBody *DeployContractResponseBody
// 	// json.Unmarshal(body, &responseBody)
// 	// return responseBody
// }

func PublishABI(ethconnectUrl string, contract *Contract) (*PublishAbiResponseBody, error) {
	u, _ := url.Parse(ethconnectUrl)
	u, _ = u.Parse("abis")
	requestUrl := u.String()

	fmt.Println("publish " + requestUrl)

	abi, _ := json.Marshal(contract.ABI)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormField("abi")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, bytes.NewReader(abi))
	if err != nil {
		return nil, err
	}
	fw, err = writer.CreateFormField("bytecode")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(fw, strings.NewReader(contract.Bytecode))
	if err != nil {
		return nil, err
	}
	writer.Close()

	req, _ := http.NewRequest("POST", requestUrl, bytes.NewReader(body.Bytes()))
	req.Header.Add("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// fmt.Println("response Status:", resp.Status)
	// fmt.Println("response Headers:", resp.Header)
	responseBody, _ := ioutil.ReadAll(resp.Body)
	var publishAbiResponse *PublishAbiResponseBody
	json.Unmarshal(responseBody, &publishAbiResponse)
	return publishAbiResponse, nil
}

func DeployContract(ethconnectUrl string, abiId string, fromAddress string, params map[string]string, registeredName string) (*DeployContractResponseBody, error) {
	u, _ := url.Parse(ethconnectUrl)
	u, _ = u.Parse(path.Join("abis", abiId))
	requestUrl := u.String()

	fmt.Println("deploy " + requestUrl)

	requestBody, _ := json.Marshal(params)

	req, _ := http.NewRequest("POST", requestUrl, bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-kaleido-from", fromAddress)
	req.Header.Set("x-kaleido-sync", "true")
	if registeredName != "" {
		req.Header.Set("x-kaleido-register", registeredName)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)
	var deployContractResponse *DeployContractResponseBody
	json.Unmarshal(responseBody, &deployContractResponse)
	return deployContractResponse, nil
}

func RegisterContract(ethconnectUrl string, abiId string, contractAddress string, fromAddress string, registeredName string, params map[string]string) (*RegisterResponseBody, error) {
	u, _ := url.Parse(ethconnectUrl)
	u, _ = u.Parse(path.Join("abis", abiId, contractAddress))
	requestUrl := u.String()

	fmt.Println("register " + requestUrl)

	// requestBody, _ := json.Marshal(params)

	req, _ := http.NewRequest("POST", requestUrl, bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-kaleido-sync", "true")
	// req.Header.Set("x-kaleido-from", fromAddress)
	req.Header.Set("x-kaleido-register", registeredName)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)
	var registerResponseBody *RegisterResponseBody
	json.Unmarshal(responseBody, &registerResponseBody)
	return registerResponseBody, nil
}

func GetReply(ethconnectUrl string, transactionId string) {

	req, _ := http.NewRequest("GET", ethconnectUrl+"/replies", nil)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("")
	fmt.Println(string(body))
}
