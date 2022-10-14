package main

import (
	"errors"
	"io/ioutil"
	"net/http"
)

type HttpClient struct {
	client *http.Client
}

func NewAppchainClient() *AppchainClient {
	var httpClient AppchainClient = &HttpClient{
		client: &http.Client{},
	}
	return &httpClient
}

const DATA_URL = "http://localhost:3005/getCustomPackage"

func (httpClient *HttpClient) GetValue(consumerPackageId string) (string, error){
	req, err := http.NewRequest("GET", DATA_URL + "?consumerPackageId=" + consumerPackageId, nil)
	if err != nil {
		return "", err
	}
	res, err := httpClient.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (httpClient *HttpClient) SetValue(key string, value string) error {

	return errors.New("HTTP尚未实现设置数据的方法")
}