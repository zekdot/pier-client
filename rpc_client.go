package main

import (
	"encoding/json"
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
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
		"PollingHelper",
		[]string{string(mStr)},
	}
	if err = rpcClient.client.Call("Service.PollingHelper", reqArgs, &reply); err != nil {
		return nil, err
	}
	if reply != "[]" {
		logger.Debug("in a fetch--- " + reply)
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

func MultiRead(threadNum int, keys []string, sourceChainID, sequenceNum, targetCID, isReqStr string, client *rpc.Client) (string, error) {
	//threadNum := 10
	ch := make(chan string, len(keys))
	//done := make(chan bool, 5)
	res := make([][]string, 0)
	var wg sync.WaitGroup
	if len(keys) < threadNum {
		threadNum = len(keys)
	}
	wg.Add(threadNum)
	for j := 0; j < threadNum; j ++ {
		go func(index uint64) {
			var key string
			for true {
				select {
				case key =<- ch:
					if key == "done" {
						//fmt.Println("进程" + strconv.FormatUint(index, 10) + "结束")
						wg.Done()
						return
					}
					reqArgs := ReqArgs{
						"InvokeInterchainHelper",
						[]string{sourceChainID, sequenceNum, targetCID, isReqStr, "interchainGet", key, ""},
					}
					var reply string

					if err := client.Call("Service.InvokeInterchainHelper", reqArgs, &reply); err != nil {
						logger.Error(err.Error())
					}
					res = append(res, []string{key, reply})
					//fmt.Println("处理" + key)
					//time.Sleep(1 * time.Millisecond)
				}
			}

		}(uint64(j))
	}
	for _, key := range keys {
		ch <- key
	}
	for j := 0; j < threadNum; j ++ {
		ch <- "done"
	}
	wg.Wait()
	resStr, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(resStr), nil
}

func (rpcClient *RpcClient) InvokeInterchainHelper(writeC chan ReqArgs, sourceChainID, sequenceNum, targetCID string, isReq bool, bizCallData []byte) (string, error) {

	isReqStr := strconv.FormatBool(isReq)

	callFunc := &CallFunc{}
	if err := json.Unmarshal(bizCallData, &callFunc); err != nil {
		return "", fmt.Errorf("unmarshal call func failed for %s", string(bizCallData))
	}
	//funcName := callFunc.Func
	//key := string(callFunc.Args[0])
	var value = ""

	var valuePart = make([]string, 0)
	// If this is a callback, jump the first callback parameter
	if callFunc.Func == "bundleResponse" {
		callFunc.Args = callFunc.Args[1:]
	}
	for _, arg := range callFunc.Args {
		valuePart = append(valuePart, string(arg))
	}
	value = strings.Join(valuePart, ",")

	if callFunc.Func == "bundleResponse" {
		// concat all args
		kvpairs := make([][]string, 0)
		err := json.Unmarshal([]byte(value), &kvpairs)
		if err != nil {
			return "", err
		}
		for _, kv := range kvpairs {
			reqArgs := ReqArgs {
				"InvokeInterchainHelper",
				[]string{sourceChainID, sequenceNum, targetCID, isReqStr, "interchainSet", kv[0], kv[1]},
			}
			writeC <- reqArgs
		}


	}

	if callFunc.Func == "bundleRequest" {
		keys := make([]string, 0)
		if err := json.Unmarshal([]byte(value), &keys); err != nil {
			return "", err
		}

		reply, err := MultiRead(10, keys, sourceChainID, sequenceNum, targetCID, isReqStr, rpcClient.client)
		if err != nil {
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