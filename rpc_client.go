package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/log"
	"net/rpc"
	"strconv"
	"strings"
)

const (
	delimiter = "&"
	outmeta = "outter-meta"
	inmeta = "inner-meta"
	callbackmeta = "callback-meta"
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

func (rpcClient *RpcClient) InterchainSet(args[] string) error {
	if len(args) < 5 {
		return fmt.Errorf("incorrect number of arguments, expecting 5")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]
	data := args[4]

	if err := rpcClient.checkIndex(sourceChainID, sequenceNum, callbackmeta); err != nil {
		return err
	}

	idx, err := strconv.ParseUint(sequenceNum, 10, 64)
	if err != nil {
		return err
	}

	if err := rpcClient.markCallbackCounter(sourceChainID, idx); err != nil {
		return err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}


	var reply string
	reqArgs := ReqArgs{
		"set",
		[]string{key, data},
	}
	err = rpcClient.client.Call("Service.SetValue", reqArgs, &reply)
	if err != nil {
		return err
	}

	return nil
}

func (rpcClient *RpcClient) InterchainGet(args[] string) (string, error) {

	if len(args) < 4 {
		return "", fmt.Errorf("incorrect number of arguments, expecting 4")
	}
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	key := args[3]

	if err := rpcClient.checkIndex(sourceChainID, sequenceNum, "inner-meta"); err != nil {
		return "", err
	}

	if err := rpcClient.markInCounter(sourceChainID); err != nil {
		return "", err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return "", fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	// args[0]: key
	value, err := rpcClient.GetData(key)
	if err != nil {
		return "", err
	}

	inKey := inMsgKey(sourceChainID, sequenceNum)
	if err := rpcClient.SetData(inKey, value); err != nil {
		return "", err
	}

	return value, nil
}

/************* start of cross-chain related method area **************/

func (rpcClient *RpcClient) EmitInterchainEvent(args []string) error {
	if len(args) != 8 {
		return fmt.Errorf("incorrect number of arguments, expecting 8")
	}
	if len(args[0]) == 0 || len(args[1]) == 0{
		return fmt.Errorf("incorrect nil destination appchain id or destination contract address")
	}

	destChainID := args[0]
	outMeta, err := rpcClient.GetOuterMeta()

	if _, ok := outMeta[destChainID]; !ok {
		outMeta[destChainID] = 0
	}

	//cid, err := getChaincodeID(stub)
	cid := "mychannel&broker"
	if err != nil {
		return err
	}

	tx := &Event{
		Index:         outMeta[destChainID] + 1,
		DstChainID:    destChainID,
		SrcContractID: cid,
		DstContractID: args[1],
		Func:          args[2],
		Args:          args[3],
		Callback:      args[4],
		Argscb:        args[5],
		Rollback:      args[6],
		Argsrb:        args[7],
	}

	outMeta[tx.DstChainID]++

	txValue, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	// persist out message
	outKey := outMsgKey(tx.DstChainID, strconv.FormatUint(tx.Index, 10))
	if err := rpcClient.SetData(outKey, string(txValue)); err != nil {
		return err
	}

	// save outter-meta
	metaStr, err := json.Marshal(outMeta)
	if err != nil {
		return err
	}
	if err := rpcClient.SetData(outmeta, string(metaStr)); err != nil {
		return err
	}
	return nil
}

func (rpcClient *RpcClient) InvokeIndexUpdate(sourceChainID, sequenceNum string, isReq bool) error {
	if err := rpcClient.updateIndex(sourceChainID, sequenceNum, isReq); err != nil {
		return err
	}
	return nil
}

// TODO finish it
func (rpcClient *RpcClient) InvokeInterchain(sourceChainID, sequenceNum, targetCID string, isReq bool, bizCallData []byte) (string, error) {

	if err := rpcClient.updateIndex(sourceChainID, sequenceNum, isReq); err != nil {
		return "", err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return "", fmt.Errorf("Target chaincode id %s is not valid", targetCID)
	}

	callFunc := &CallFunc{}
	if err := json.Unmarshal(bizCallData, callFunc); err != nil {
		return "", fmt.Errorf("unmarshal call func failed for %s", string(bizCallData))
	}

	// TODO use callFunc to call related method, Still don't know what Func and Args will be.. emmm
	log.Infof("call func %s", callFunc.Func)
	for arg, idx := range callFunc.Args {
		log.Infof("\targ%d is %s", idx, string(arg))
	}
	//var ccArgs [][]byte
	//ccArgs = append(ccArgs, []byte(callFunc.Func))
	//ccArgs = append(ccArgs, callFunc.Args...)
	//response := stub.InvokeChaincode(splitedCID[1], ccArgs, splitedCID[0])
	//if response.Status != shim.OK {
	//	return errorResponse(fmt.Sprintf("invoke chaincode '%s' function %s err: %s", splitedCID[1], callFunc.Func, response.Message))
	//}
	response := "success"

	inKey := inMsgKey(sourceChainID, sequenceNum)
	value, err := json.Marshal(response)
	err = rpcClient.SetData(inKey, string(value))
	if err != nil {
		return "", err
	}

	return response, nil
}

func (rpcClient *RpcClient) Polling(m map[string]uint64) ([]*Event, error) {
	outMeta, err := rpcClient.GetOuterMeta()
	if err != nil {
		return nil, err
	}
	events := make([]*Event, 0)
	for addr, idx := range outMeta {
		startPos, ok := m[addr]
		if !ok {
			startPos = 0
		}
		for i := startPos + 1; i <= idx; i++ {
			event, _ := rpcClient.GetOutMessage(addr, i)

			events = append(events, event)
		}
	}
	return events, nil
}

/************* end of cross-chain related method area **************/

/************* start of index-helper related method area **************/

func (rpcClient *RpcClient) markInCounter(from string) error {
	inMeta, err := rpcClient.GetInnerMeta()
	if err != nil {
		return err
	}

	inMeta[from]++
	metaStr, err := json.Marshal(inMeta)
	if err != nil {
		return err
	}
	err = rpcClient.SetData(inmeta, string(metaStr))
	return err
}

func (rpcClient *RpcClient) markCallbackCounter(from string, index uint64) error {
	meta, err := rpcClient.GetCallbackMeta()
	if err != nil {
		return err
	}

	meta[from] = index
	metaStr, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	err = rpcClient.SetData(callbackmeta, string(metaStr))
	return err
}

func (rpcClient *RpcClient) checkIndex(addr string, index string, metaName string) error {
	idx, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return err
	}
	meta, err := rpcClient.getMeta(metaName)  //broker.getMap(state, metaName)
	if err != nil {
		return err
	}
	if idx != meta[addr] + 1 {
		return fmt.Errorf("incorrect index, expect %d", meta[addr]+1)
	}
	return nil
}

func (rpcClient *RpcClient) updateIndex(sourceChainID, sequenceNum string, isReq bool) error {
	if isReq {
		if err := rpcClient.checkIndex(sourceChainID, sequenceNum, inmeta); err != nil {
			return err
		}
		if err := rpcClient.markInCounter(sourceChainID); err != nil {
			return err
		}
	} else {
		if err := rpcClient.checkIndex(sourceChainID, sequenceNum, callbackmeta); err != nil {
			return err
		}

		idx, err := strconv.ParseUint(sequenceNum, 10, 64)
		if err != nil {
			return err
		}
		if err := rpcClient.markCallbackCounter(sourceChainID, idx); err != nil {
			return err
		}
	}

	return nil
}

/************* end of index-helper related method area **************/

/************* start of meta-data related method area **************/

func (rpcClient *RpcClient) Init() error {
	err := rpcClient.SetData(inmeta, "{}")
	if err != nil {
		return err
	}
	err = rpcClient.SetData(outmeta, "{}")
	if err != nil {
		return err
	}
	err = rpcClient.SetData(callbackmeta, "{}")
	if err != nil {
		return err
	}
	return nil
}

// ToChaincodeArgs converts string args to []byte args
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

func (rpcClient *RpcClient) GetInMessage(sourceChainID string, sequenceNum uint64)([][]byte, error) {
	key := inMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reply, err := rpcClient.GetData(key)
	if err != nil {
		return nil, err
	}
	results := strings.Split(reply, ",")
	return toChaincodeArgs(results...), nil
}

func (rpcClient *RpcClient) GetOutMessage(sourceChainID string, sequenceNum uint64)(*Event, error) {
	key := outMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reply, err := rpcClient.GetData(key)
	if err != nil {
		return nil, err
	}
	ret := &Event{}
	if err := json.Unmarshal([]byte(reply), ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (rpcClient *RpcClient) getMeta(key string) (map[string]uint64, error) {
	reply, err := rpcClient.GetData(key)
	if err != nil {
		return nil, err
	}
	outMeta := make(map[string]uint64)
	err = json.Unmarshal([]byte(reply), &outMeta)
	if err != nil {
		return nil, err
	}
	return outMeta, nil
}

func (rpcClient *RpcClient) GetInnerMeta() (map[string]uint64, error) {
	return rpcClient.getMeta("inner-meta")
}

func (rpcClient *RpcClient) GetOuterMeta() (map[string]uint64, error) {
	return rpcClient.getMeta("outter-meta")
}

func (rpcClient *RpcClient) GetCallbackMeta() (map[string]uint64, error) {
	return rpcClient.getMeta("callback-meta")
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func inMsgKey(to string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", to, idx)
}

/************* end of meta-data related method area **************/


/************* start of called-by-others base method area **************/

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

/************* end of called-by-others base method area **************/