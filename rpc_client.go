package main

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
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

func (rpcClient *RpcClient) GetMeta(key string) (map[string]uint64, error) {
	var reply string
	reqArgs := ReqArgs{
		"getMeta",
		[]string{key},
	}
	err := rpcClient.client.Call("Service.GetMeta", reqArgs, &reply)
	if err != nil {
		return nil, err
	}
	res := make(map[string]uint64)
	json.Unmarshal([]byte(reply), &res)
	return res, nil
}


func (rpcClient *RpcClient) PollingHelper(m map[string]uint64) ([]*Event, error) {

	var reply string
	mStr, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	reqArgs := ReqArgs{
		"getMeta",
		[]string{string(mStr)},
	}
	if err = rpcClient.client.Call("Service.PollingHelper", reqArgs, &reply); err != nil {
		return nil, err
	}
	res := make([]*Event, 0)
	json.Unmarshal([]byte(reply), &res)
	return res, nil
}

func inMsgKey(to string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", to, idx)
}

// ToChaincodeArgs converts string args to []byte args
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func (rpcClient *RpcClient) GetInMessageHelper(sourceChainID string, sequenceNum uint64)([][]byte, error) {

	var reply string
	key := inMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reqArgs := ReqArgs{
		"GetInMessageStrByKey",
		[]string{string(key)},
	}
	if err := rpcClient.client.Call("Service.GetInMessageStrByKey", reqArgs, &reply); err != nil {
		return nil, err
	}
	results := []string{"true"}
	results = append(results, strings.Split(string(reply), ",")...)
	return toChaincodeArgs(results...), nil
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func (rpcClient *RpcClient) GetOutMessageHelper(sourceChainID string, sequenceNum uint64)(*Event, error) {

	var reply string
	key := outMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))

	reqArgs := ReqArgs{
		"GetOutMessageByKey",
		[]string{key},
	}
	if err := rpcClient.client.Call("Service.GetOutMessageStrByKey", reqArgs, &reply); err != nil {
		return nil, err
	}
	res := &Event{}
	json.Unmarshal([]byte(reply), &res)
	return res, nil
}

func (rpcClient *RpcClient) InvokeInterchainHelper(writeC chan ReqArgs, sourceChainID, sequenceNum, targetCID string, isReq bool, bizCallData []byte) (string, error) {

	isReqStr := strconv.FormatBool(isReq)

	callFunc := &CallFunc{}
	if err := json.Unmarshal(bizCallData, &callFunc); err != nil {
		return "", fmt.Errorf("unmarshal call func failed for %s", string(bizCallData))
	}
	funcName := callFunc.Func
	key := string(callFunc.Args[0])
	var value = ""



	if callFunc.Func == "interchainSet" {
		// concat all args except first arg with comma
		var valuePart = make([]string, 0)
		for _, arg := range callFunc.Args[1:] {
			valuePart = append(valuePart, string(arg))
		}
		value = strings.Join(valuePart, ",")
		//return value, nil
		reqArgs := ReqArgs{
			"InvokeInterchainHelper",
			[]string{sourceChainID, sequenceNum, targetCID, isReqStr, funcName, key, value},
		}
		writeC <- reqArgs
	}

	if callFunc.Func == "interchainGet" {
		reqArgs := ReqArgs{
			"InvokeInterchainHelper",
			[]string{sourceChainID, sequenceNum, targetCID, isReqStr, funcName, key, value},
		}
		var reply string
		if err := rpcClient.client.Call("Service.InvokeInterchainHelper", reqArgs, &reply); err != nil {
			return "", err
		}
		return reply, nil
	}
	//var reply string

	return "success", nil
	//if err := rpcClient.client.Call("Service.InvokeInterchainHelper", reqArgs, &reply); err != nil {
	//	return "", err
	//}
	//return reply, nil
}

func (rpcClient *RpcClient) UpdateIndexHelper(sourceChainID, sequenceNum, isReq string)  error {

	var reply string

	reqArgs := ReqArgs{
		"UpdateIndexHelper",
		[]string{sourceChainID, sequenceNum, isReq},
	}
	return rpcClient.client.Call("Service.UpdateIndexHelper", reqArgs, &reply)
}