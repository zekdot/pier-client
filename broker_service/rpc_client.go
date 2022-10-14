package main

import (
	"net/rpc"
)

type AppchainClient struct {
	client *rpc.Client
}

type AppchainReqArgs struct {
	Args []string
}

func NewAppchainClient(address string) (*AppchainClient, error) {
	rpcClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, err
	}
	return &AppchainClient{
		client: rpcClient,
	}, nil
}

func (rpcClient *AppchainClient) GetValue(key string) (string, error) {
	var reply string
	appchainReqArgs := AppchainReqArgs{
		[]string{key},
	}
	err := rpcClient.client.Call("Service.GetValue", appchainReqArgs, &reply)
	if err != nil {
		return "", err
	}
	return reply, nil
}

func (rpcClient *AppchainClient) SetValue(key string, value string) error {
	var reply string
	appchainReqArgs := AppchainReqArgs{
		[]string{key, value},
	}
	err := rpcClient.client.Call("Service.SetValue", appchainReqArgs, &reply)
	if err != nil {
		return err
	}
	return nil
}
