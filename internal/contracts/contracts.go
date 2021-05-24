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

type DeployContractResponseBody struct {
	ContractAddress string `json:"contractAddress,omitempty"`
}

type RegisterResponseBody struct {
	Created      string `json:"created,omitempty"`
	Address      string `json:"string,omitempty"`
	Path         string `json:"path,omitempty"`
	ABI          string `json:"ABI,omitempty"`
	OpenAPI      string `json:"openapi,omitempty"`
	RegisteredAs string `json:"registeredAs,omitempty"`
}

func ReadCompiledContract(filePath string) *Contract {
	d, _ := ioutil.ReadFile(filePath)
	var contract *Contract
	err := json.Unmarshal(d, &contract)
	if err != nil {
		fmt.Println(err.Error())
	}
	return contract
}

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

	req, _ := http.NewRequest("POST", requestUrl, bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-kaleido-sync", "true")
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
