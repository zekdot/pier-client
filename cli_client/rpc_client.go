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
const (
	delimiter = "&"
	outmeta = "outter-meta"
	inmeta = "inner-meta"
	callbackmeta = "callback-meta"
)

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
	Argscb        string `json:"argscb"`
	Rollback      string `json:"rollback"`
	Argsrb        string `json:"argsrb"`
	Proof         []byte `json:"proof"`
	Extra         []byte `json:"extra"`
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

func (rpcClient *RpcClient) GetOuterMeta() (map[string]uint64, error) {
	return rpcClient.getMeta(outmeta)
}

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

func (rpcClient *RpcClient) InterchainGet(toId string, contractId string, key string) error {
	args := []string{toId, contractId, "interchainGet", key, "interchainSet", key, "", ""}
	return rpcClient.EmitInterchainEvent(args)
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}