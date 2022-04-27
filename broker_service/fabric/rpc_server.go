package main

import (
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"os"

	//"github.com/wonderivan/logger"

	//"github.com/hashicorp/go-hclog"
	//"os"
)

var (
	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "performance",
		Output: os.Stdout,
		Level:  hclog.Trace,
	})
)
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
type Service struct {
	broker *BrokerClient
}


type ReqArgs struct {
	FuncName string
	Args []string
}

func NewService(broker *BrokerClient) *Service {
	return &Service{
		broker: broker,
	}
}

// send transaction and don't need result
func (s *Service) SetValue(req *ReqArgs, reply *string) error{

	broker := s.broker
	args := req.Args
	// if this is a outer cross-chain request
	if len(args[0]) > 7 && args[0][0:7] == "out-msg" {
		evt := &Event{}
		_ = json.Unmarshal([]byte(args[1]), evt)
		logger.Info("s1:key-" + evt.Args +" save cross-chain request to ledger")
	}
	//fmt.Printf("set %s to %s\n", args[0], args[1])
	err := broker.setValue(args[0], args[1])
	if len(args[0]) > 7 && args[0][0:7] == "out-msg" {
		evt := &Event{}
		_ = json.Unmarshal([]byte(args[1]), evt)
		logger.Info("s2:key-" + evt.Args +" have saved cross-chain request to ledger")
	}
	return err
}

// query transaction and need result
func (s *Service) GetValue(req *ReqArgs, reply *string) error{
	broker := s.broker
	args := req.Args
	//fmt.Printf("get value of %s\n", args[0])
	res, err := broker.getValue(args[0])
	*reply = string(res)
	return err
}