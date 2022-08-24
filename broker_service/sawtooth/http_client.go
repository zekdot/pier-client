package main

import (
	"io/ioutil"
	"net/http"
)

type HttpClient struct {
	client *http.Client
}

func NewHttpClient() *HttpClient{
	return &HttpClient{
		client: &http.Client{},
	}
}

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