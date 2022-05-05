package main

import (
	"encoding/json"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
)

type DB struct {
	db *leveldb.DB
}

func NewDB(path string) (*DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return &DB{
		db: db,
	}, nil
}


const (
	delimiter = "&"
	outmeta = "outter-meta"
	inmeta = "inner-meta"
	callbackmeta = "callback-meta"
)

func (db *DB) Init() error {
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
	return db.db.Write(batch, nil)
}

func (db *DB) InterchainGet(destChainID, contractId, key string) error {

	outMetaBytes, err := db.db.Get([]byte(outmeta), nil)// rpcClient.GetOuterMeta()
	outMeta := make(map[string]uint64)
	if err = json.Unmarshal(outMetaBytes, &outMeta); err != nil {
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

	// use batch update
	batch := new(leveldb.Batch)
	batch.Put([]byte(outKey), txValue)
	batch.Put([]byte(outmeta), metaBytes)

	return db.db.Write(batch, nil)
}

func (db *DB) PollingEvents(m map[string]uint64) ([]*Event, error) {
	//outMeta, err := c.getMeta(outmeta)
	outMetaBytes, err := db.db.Get([]byte(outmeta), nil)
	outMeta := make(map[string]uint64)
	json.Unmarshal(outMetaBytes, &outMeta)


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
			event, _ := db.GetOutMessageHelper(addr, i)

			events = append(events, event)
		}
	}
	return events, nil
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func (db *DB) GetOutMessageHelper(sourceChainID string, sequenceNum uint64)(*Event, error) {
	key := outMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reply, err := db.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	ret := &Event{}
	if err := json.Unmarshal(reply, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func (db *DB) GetOutMessageStrByKey(key string)(string, error) {
	//key := outMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reply, err := db.db.Get([]byte(key), nil)
	if err != nil {
		return "", err
	}
	return string(reply), nil
}

func (db *DB) GetInMessageStrByKey(key string)(string, error) {
	//key := inMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reply, err := db.db.Get([]byte(key), nil)
	if err != nil {
		return "", err
	}
	return string(reply), nil
}

func inMsgKey(to string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", to, idx)
}

func (db *DB) GetMeta(key string) (map[string]uint64, error) {
	value, err := db.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	meta := make(map[string]uint64)
	err = json.Unmarshal(value, &meta)
	if err != nil {
		return nil, err
	}
	return meta, nil
}

func (db *DB) setMeta(key string, value map[string]uint64) error {
	bvalue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return db.db.Put([]byte(key), bvalue, nil)
}

func (db *DB) GetMetaStr(key string) (string, error) {
	value, err := db.db.Get([]byte(key), nil)
	if err != nil {
		return "", err
	}
	return string(value), nil
}

func (db *DB) InvokeInterchainHelper(sourceChainID, sequenceNum, targetCID, isReq, funcName, key, value string, client *BrokerClient) (string, error) {

	if err := db.UpdateIndex(sourceChainID, sequenceNum, isReq); err != nil {
		return "", err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return "", fmt.Errorf("target chaincode id %s is not valid", targetCID)
	}

	var res = "unknown"
	var valueBytes []byte
	var err error
	if funcName == "interchainGet"  {
		logger.Info("s6:key-" + key + " try to get value from sawtooth")
		valueBytes, err = client.getValue(key)
		res = string(valueBytes)
		logger.Info("s7:key-" + key + " get value from sawtooth successfully")
		if err != nil {
			return "", err
		}
		inKey := inMsgKey(sourceChainID, sequenceNum)
		err = db.db.Put([]byte(inKey), []byte(value), nil)
		if err != nil {
			return "", err
		}
	} else if funcName == "interchainSet" {
		logger.Info("s4:key-" + key + " submit interchainSet request")
		err = client.setValue(key, value)
		if err != nil {
			return "", err
		}
	}
	return res, nil
}

func (db *DB) markInCounter(from string) error {
	inMeta, err := db.GetMeta(inmeta) // rpcClient.GetInnerMeta()
	if err != nil {
		return err
	}
	inMeta[from]++
	return db.setMeta(inmeta, inMeta)
}

func (db *DB) markCallbackCounter(from string, index uint64) error {
	meta, err := db.GetMeta(callbackmeta) // rpcClient.GetCallbackMeta()
	if err != nil {
		return err
	}

	meta[from] = index
	return db.setMeta(callbackmeta, meta)
}

func (db *DB) checkIndex(addr string, index string, metaName string) error {
	idx, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return err
	}
	meta, err := db.GetMeta(metaName) // rpcClient.getMeta(metaName)  //broker.getMap(state, metaName)
	if err != nil {
		return err
	}
	if idx != meta[addr] + 1 {
		return fmt.Errorf("incorrect index, expect %d", meta[addr]+1)
	}
	return nil
}

func (db *DB) UpdateIndex(sourceChainID, sequenceNum, isReq string) error {
	if isReq == "true" {
		if err := db.checkIndex(sourceChainID, sequenceNum, inmeta); err != nil {
			return err
		}
		if err := db.markInCounter(sourceChainID); err != nil {
			return err
		}
	} else {
		if err := db.checkIndex(sourceChainID, sequenceNum, callbackmeta); err != nil {
			return err
		}

		idx, err := strconv.ParseUint(sequenceNum, 10, 64)
		if err != nil {
			return err
		}
		if err := db.markCallbackCounter(sourceChainID, idx); err != nil {
			return err
		}
	}

	return nil
}
