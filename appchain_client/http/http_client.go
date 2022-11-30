package main

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"encoding/json"
)

type HttpClient struct {
	client *http.Client
	token string
}

func NewAppchainClient() *AppchainClient {
	var httpClient AppchainClient = &HttpClient{
		client: &http.Client{},
		token: "",
	}
	return &httpClient
}

// try to login with configured username and password
func (httpClient *HttpClient) Login() (string, error) {
	req, err := http.PostForm(TOKEN_URL, url.Values{
		"email":{ EMAIL },
		"password":{PASSWORD},
	})

	if err != nil {
		return "", err
	}
	// res, err := httpClient.client.Do(req)
	if err != nil {
		return "", err
	}
	defer req.Body.Close()

	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return "", err
	}
	var dataMap map[string]string
	err = json.Unmarshal(body, &dataMap)
	if err != nil {
		return "", err
	}
	return dataMap["token"], nil
}

func (httpClient *HttpClient) GetValue(consumerPackageId string) (string, error){
	req, err := http.NewRequest("GET", DATA_URL + "/" + consumerPackageId, nil)
	req.Header.Set("Authorization", "Bearer " + httpClient.token)
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
	if string(body) == "{\"statusCode\":401,\"message\":\"Unauthorized\"}" {
		var err error
		httpClient.token, err = httpClient.Login()
		if err != nil {
			return "", nil
		}
		req, _ := http.NewRequest("GET", DATA_URL + "/" + consumerPackageId, nil)
		req.Header.Set("Authorization", "Bearer " + httpClient.token)
		res, _ := httpClient.client.Do(req)
		body, err = ioutil.ReadAll(res.Body)
	}
	return string(body), nil
}

func (httpClient *HttpClient) SetValue(key string, value string) error {

	return errors.New("HTTP尚未实现设置数据的方法")
}
