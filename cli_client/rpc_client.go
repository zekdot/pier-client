package main

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"strconv"
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
		[]string{"inner-meta", "{}"},
	}
	err := rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	reqArgs = ReqArgs{
		"init",
		[]string{"outter-meta", "{}"},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	reqArgs = ReqArgs{
		"init",
		[]string{"callback-meta", "{}"},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
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

type Event struct {
	Index         uint64 `json:"index"`
	DstChainID    string `json:"dst_chain_id"`
	SrcContractID string `json:"src_contract_id"`
	DstContractID string `json:"dst_contract_id"`
	Func          string `json:"func"`
	Args          string `json:"args"`
	Callback      string `json:"callback"`
}

func (rpcClient *RpcClient) InterchainGet(toId string, contractId string, key string) error {
	cid := "mychannel&data_swapper"
	//destChainId := toId
	var reply string
	reqArgs := ReqArgs{
		"get",
		[]string{"outter-meta"},
	}
	err := rpcClient.client.Call("Service.GetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	outMeta := make(map[string]uint64)
	err = json.Unmarshal([]byte(reply), &outMeta)
	if err != nil {
		return err
	}
	if _, ok := outMeta[toId]; !ok {
		outMeta[toId] = 0
	}
	tx := &Event{
		Index:         outMeta[toId] + 1,
		DstChainID:    toId,
		SrcContractID: cid,
		DstContractID: contractId,
		Func:          "interchainGet",
		Args:          key,
		Callback:      "interchainSet",
	}
	txValue, err := json.Marshal(tx)
	outKey := outMsgKey(tx.DstChainID, strconv.FormatUint(tx.Index, 10))
	reqArgs = ReqArgs{
		"set",
		[]string{outKey, string(txValue)},
	}
	if err != nil {
		return err
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply);
	if err != nil {
		return err
	}
	outMeta[toId] ++
	outMetaBytes, err := json.Marshal(outMeta)
	if err != nil {
		return err
	}
	reqArgs = ReqArgs{
		"set",
		[]string{"outter-meta", string(outMetaBytes)},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}
	return nil
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}