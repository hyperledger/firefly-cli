package contracts

import (
	"bytes"
	"encoding/json"
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

func ReadCompiledContract(filePath string) (*Contract, error) {
	d, _ := ioutil.ReadFile(filePath)
	var contract *Contract
	err := json.Unmarshal(d, &contract)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

func PublishABI(ethconnectUrl string, contract *Contract) (*PublishAbiResponseBody, error) {
	u, err := url.Parse(ethconnectUrl)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse("abis")
	if err != nil {
		return nil, err
	}
	requestUrl := u.String()
	abi, err := json.Marshal(contract.ABI)
	if err != nil {
		return nil, err
	}
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	fw, err := writer.CreateFormField("abi")
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(fw, bytes.NewReader(abi)); err != nil {
		return nil, err
	}
	fw, err = writer.CreateFormField("bytecode")
	if err != nil {
		return nil, err
	}
	if _, err = io.Copy(fw, strings.NewReader(contract.Bytecode)); err != nil {
		return nil, err
	}
	writer.Close()
	req, err := http.NewRequest("POST", requestUrl, bytes.NewReader(body.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", writer.FormDataContentType())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var publishAbiResponse *PublishAbiResponseBody
	json.Unmarshal(responseBody, &publishAbiResponse)
	return publishAbiResponse, nil
}

func DeployContract(ethconnectUrl string, abiId string, fromAddress string, params map[string]string, registeredName string) (*DeployContractResponseBody, error) {
	u, err := url.Parse(ethconnectUrl)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("abis", abiId))
	if err != nil {
		return nil, err
	}
	requestUrl := u.String()
	requestBody, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
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
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var deployContractResponse *DeployContractResponseBody
	json.Unmarshal(responseBody, &deployContractResponse)
	return deployContractResponse, nil
}

func RegisterContract(ethconnectUrl string, abiId string, contractAddress string, fromAddress string, registeredName string, params map[string]string) (*RegisterResponseBody, error) {
	u, err := url.Parse(ethconnectUrl)
	if err != nil {
		return nil, err
	}
	u, err = u.Parse(path.Join("abis", abiId, contractAddress))
	if err != nil {
		return nil, err
	}
	requestUrl := u.String()
	req, err := http.NewRequest("POST", requestUrl, bytes.NewBuffer(nil))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-kaleido-sync", "true")
	req.Header.Set("x-kaleido-register", registeredName)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var registerResponseBody *RegisterResponseBody
	json.Unmarshal(responseBody, &registerResponseBody)
	return registerResponseBody, nil
}
