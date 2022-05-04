package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/syndtr/goleveldb/leveldb"
	"os"
	"strconv"
)
const (
	outmeta = "outter-meta"
	inmeta = "inner-meta"
	callbackmeta = "callback-meta"
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
	db *leveldb.DB
}

func NewDB(path string) (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
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

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

// query transaction and need result
func (s *Service) Init(req *ReqArgs, reply *string) error{
	emptyMap := make(map[string]uint64)
	var emptyMapBytes []byte
	emptyMapBytes, err := json.Marshal(emptyMap)
	if err != nil {
		return err
	}
	batch := new(leveldb.Batch)
	batch.Put([]byte(inmeta), emptyMapBytes)
	batch.Put([]byte(outmeta), emptyMapBytes)
	batch.Put([]byte(callbackmeta), emptyMapBytes)
	return s.db.Write(batch, nil)
}

func (s *Service) InterchainGet(req *ReqArgs, reply *string) error {
	args := req.Args

	destChainID := args[0]
	contractId := args[1]
	key := args[2]
	outMetaBytes, err := s.db.Get([]byte(outmeta), nil)// rpcClient.GetOuterMeta()
	outMeta := make(map[string]uint64)
	if err = json.Unmarshal(outMetaBytes, outMeta); err != nil {
		return err
	}
	if _, ok := outMeta[destChainID]; !ok {
		outMeta[destChainID] = 0
	}

	cid := "mychannel&broker"
	if err != nil {
		return err
	}
	// toId, contractId, "interchainGet", key, "interchainSet", key, "", ""
	tx := &Event{
		Index:         outMeta[destChainID] + 1,
		DstChainID:    destChainID,
		SrcContractID: cid,
		DstContractID: contractId,
		Func:          "interchainGet",
		Args:          key,
		Callback:      "interchainSet",
		Argscb:        key,
		Rollback:      "",
		Argsrb:        "",
	}

	outMeta[tx.DstChainID]++

	txValue, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	// persist out message
	outKey := outMsgKey(tx.DstChainID, strconv.FormatUint(tx.Index, 10))
	metaBytes, err := json.Marshal(outMeta)

	logger.Info("s1:key-" + key +" save cross-chain request to ledger")

	// use batch update
	batch := new(leveldb.Batch)
	batch.Put([]byte(outKey), txValue)
	batch.Put([]byte(outmeta), metaBytes)
	defer logger.Info("s2:key-" + key +" have saved cross-chain request to ledger")

	return s.db.Write(batch, nil)
}