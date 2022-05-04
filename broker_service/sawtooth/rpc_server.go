package main

import (
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"os"
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
	db *DB
}

type ReqArgs struct {
	FuncName string
	Args []string
}

func NewService(broker *BrokerClient) *Service {
	db, err := NewDB(DB_PATH)
	if err != nil {
		panic(err)
	}
	return &Service{
		broker: broker,
		db: db,
	}
}

// send transaction and don't need result
func (s *Service) SetValue(req *ReqArgs, reply *string) error{
	broker := s.broker
	args := req.Args
	err := broker.setValue(args[0], args[1])
	return err
}

// query transaction and need result
func (s *Service) GetValue(req *ReqArgs, reply *string) error{
	broker := s.broker
	args := req.Args
	res, err := broker.getValue(args[0])
	*reply = string(res)
	return err
}

// query transaction and need result
func (s *Service) Init(req *ReqArgs, reply *string) error{
	return s.db.Init()
}

func (s *Service) InterchainGet(req *ReqArgs, reply *string) error {
	args := req.Args

	destChainID := args[0]
	contractId := args[1]
	key := args[2]

	logger.Info("s1:key-" + key +" save cross-chain request to ledger")

	defer logger.Info("s2:key-" + key +" have saved cross-chain request to ledger")

	return s.db.InterchainGet(destChainID, contractId, key)
}



func (s *Service) GetMeta(req *ReqArgs, reply *string) error {
	args := req.Args

	key := args[0]

	value, err := s.db.GetMetaStr(key)
	if err != nil {
		return err
	}
	*reply = value
	return nil
}

func (s *Service) PollingHelper(req *ReqArgs, reply *string) error {
	args := req.Args

	mStr := args[0]

	m := make(map[string]uint64)
	json.Unmarshal([]byte(mStr), &m)

	evs, err := s.db.PollingEvents(m)
	if err != nil {
		return err
	}
	evsStr, err := json.Marshal(evs)
	*reply = string(evsStr)
	return err
}

func (s *Service) GetInMessageStrByKey(req *ReqArgs, reply *string) error {
	args := req.Args
	key := args[0]
	var err error
	*reply, err = s.db.GetInMessageStrByKey(key)
	return err
}

func (s *Service) GetOutMessageStrByKey(req *ReqArgs, reply *string) error {
	args := req.Args
	key := args[0]
	var err error
	*reply, err = s.db.GetOutMessageStrByKey(key)
	return err
}

func (s *Service) InvokeInterchainHelper(req *ReqArgs, reply *string) error {
	args := req.Args
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	isReq := args[3]
	funcName := args[4]
	key := args[5]
	var value string
	if len(args) > 6 {
		value = args[6]
	} else {
		value = ""
	}
	var err error
	*reply, err = s.db.InvokeInterchainHelper(sourceChainID, sequenceNum, targetCID, isReq, funcName, key, value, s.broker)
	return err
}

func (s *Service) UpdateIndexHelper(req *ReqArgs, reply *string) error {
	args := req.Args
	sourceChainID := args[0]
	sequenceNum := args[1]
	isReq := args[2]
	return s.db.UpdateIndex(sourceChainID, sequenceNum, isReq)
}
