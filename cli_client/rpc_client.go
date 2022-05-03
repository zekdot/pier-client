package main

import (
	"net/rpc"
)

type RpcClient struct {
	client *rpc.Client
}

type ReqArgs struct {
	FuncName string
	Args []string
}

func NewRpcClient(address string) (*RpcClient, error) {
	rpcClient, err := rpc.DialHTTP("tcp", address)
	if err != nil {
		return nil, err
	}
	return &RpcClient{
		client: rpcClient,
	}, nil
}

func (rpcClient *RpcClient) Init() error {
	var reply string
	reqArgs := ReqArgs{
		"init",
		[]string{},
	}
	err := rpcClient.client.Call("Service.Init", reqArgs, &reply)
	if err != nil {
		return err
	}
	return nil
}

func (rpcClient *RpcClient) GetData(key string) (string, error) {
	var reply string
	reqArgs := ReqArgs{
		"get",
		[]string{key},
	}
	err := rpcClient.client.Call("Service.GetValue", reqArgs, &reply)
	if err != nil {
		return "", err
	}
	return reply, nil
}

func (rpcClient *RpcClient) SetData(key string, value string) error {
	var reply string
	reqArgs := ReqArgs{
		"set",
		[]string{key, value},
	}
	err := rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	return nil
}

func (rpcClient *RpcClient) InterchainGet(toId string, contractId string, key string) error {
	var reply string
	reqArgs := ReqArgs{
		"",
		[]string{toId, contractId, key},
	}
	err := rpcClient.client.Call("Service.InterchainGet", reqArgs, &reply)
	if err != nil {
		return err
	}
	return nil
}