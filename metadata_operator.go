package main

import (
	"encoding/json"
	"fmt"
	"github.com/prometheus/common/log"
	"github.com/syndtr/goleveldb/leveldb"
	"strconv"
	"strings"
)

const (
	delimiter = "&"
	outmeta = "outter-meta"
	inmeta = "inner-meta"
	callbackmeta = "callback-meta"
)

func NewDB(path string) (*leveldb.DB, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// ToChaincodeArgs converts string args to []byte args
func toChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}


/************* start of cross-chain related method area **************/

func (c *Client) invokeInterchainHelper(sourceChainID, sequenceNum, targetCID string, isReq bool, bizCallData []byte) (string, error) {

	if err := c.updateIndex(sourceChainID, sequenceNum, isReq); err != nil {
		return "", err
	}

	splitedCID := strings.Split(targetCID, delimiter)
	if len(splitedCID) != 2 {
		return "", fmt.Errorf("target chaincode id %s is not valid", targetCID)
	}

	callFunc := &CallFunc{}
	if err := json.Unmarshal(bizCallData, callFunc); err != nil {
		return "", fmt.Errorf("unmarshal call func failed for %s", string(bizCallData))
	}

	// print what function will be call and what params will be
	log.Infof("call func %s", callFunc.Func)
	for idx, arg := range callFunc.Args {
		log.Infof("\targ%d is %s", idx, string(arg))
	}
	var value = "unknown"
	var err error
	if callFunc.Func == "interchainGet"  {
		logger.Info("s6:key-" + string(callFunc.Args[0]) + " try to get value from sawtooth")
		value, err = c.client.GetData(string(callFunc.Args[0]))
		logger.Info("s7:key-" + string(callFunc.Args[0]) + " get value from sawtooth successfully")
		if err != nil {
			return "", err
		}
	} else if callFunc.Func == "interchainSet" {
		logger.Info("s4:key-" + string(callFunc.Args[0]) + " submit interchainSet request")
		// still have comma problem, need to deal with, just concat all args except first arg with comma
		var valuePart = make([]string, 0)
		for _, arg := range callFunc.Args[1:] {
			valuePart = append(valuePart, string(arg))
		}
		value := strings.Join(valuePart, ",")
		err = c.client.SetData(string(callFunc.Args[0]), value)
		if err != nil {
			return "", err
		}
	}

	inKey := inMsgKey(sourceChainID, sequenceNum)
	err = c.db.Put([]byte(inKey), []byte(value), nil)
	if err != nil {
		return "", err
	}

	return value, nil
}

func (c *Client) pollingHelper(m map[string]uint64) ([]*Event, error) {
	outMeta, err := c.getMeta(outmeta)
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
			event, _ := c.getOutMessageHelper(addr, i) // rpcClient.GetOutMessage(addr, i)

			events = append(events, event)
		}
	}
	return events, nil
}

/************* end of cross-chain related method area **************/


/************* start of meta-data related method area **************/

func (c *Client) getInMessageHelper(sourceChainID string, sequenceNum uint64)([][]byte, error) {
	key := inMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reply, err := c.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	results := []string{"true"}
	//results := strings.Split(reply, ",")
	results = append(results, strings.Split(string(reply), ",")...)
	return toChaincodeArgs(results...), nil
}

func (c *Client) getOutMessageHelper(sourceChainID string, sequenceNum uint64)(*Event, error) {
	key := outMsgKey(sourceChainID, strconv.FormatUint(sequenceNum, 10))
	reply, err := c.db.Get([]byte(key), nil)
	if err != nil {
		return nil, err
	}
	ret := &Event{}
	if err := json.Unmarshal(reply, ret); err != nil {
		return nil, err
	}
	return ret, nil
}

func outMsgKey(to string, idx string) string {
	return fmt.Sprintf("out-msg-%s-%s", to, idx)
}

func inMsgKey(to string, idx string) string {
	return fmt.Sprintf("in-msg-%s-%s", to, idx)
}

func (c *Client) getMeta(key string) (map[string]uint64, error) {
	value, err := c.db.Get([]byte(key), nil)
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

func (c *Client) setMeta(key string, value map[string]uint64) error {
	bvalue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.db.Put([]byte(key), bvalue, nil)
}

/************* end of meta-data related method area **************/


/************* start of index-helper related method area **************/

func (c *Client) markInCounter(from string) error {
	inMeta, err := c.getMeta(inmeta) // rpcClient.GetInnerMeta()
	if err != nil {
		return err
	}
	inMeta[from]++
	return c.setMeta(inmeta, inMeta)
}

func (c *Client) markCallbackCounter(from string, index uint64) error {
	meta, err := c.getMeta(callbackmeta) // rpcClient.GetCallbackMeta()
	if err != nil {
		return err
	}

	meta[from] = index
	return c.setMeta(callbackmeta, meta)
}

func (c *Client) checkIndex(addr string, index string, metaName string) error {
	idx, err := strconv.ParseUint(index, 10, 64)
	if err != nil {
		return err
	}
	meta, err := c.getMeta(metaName) // rpcClient.getMeta(metaName)  //broker.getMap(state, metaName)
	if err != nil {
		return err
	}
	if idx != meta[addr] + 1 {
		return fmt.Errorf("incorrect index, expect %d", meta[addr]+1)
	}
	return nil
}

func (c *Client) updateIndex(sourceChainID, sequenceNum string, isReq bool) error {
	if isReq {
		if err := c.checkIndex(sourceChainID, sequenceNum, inmeta); err != nil {
			return err
		}
		if err := c.markInCounter(sourceChainID); err != nil {
			return err
		}
	} else {
		if err := c.checkIndex(sourceChainID, sequenceNum, callbackmeta); err != nil {
			return err
		}

		idx, err := strconv.ParseUint(sequenceNum, 10, 64)
		if err != nil {
			return err
		}
		if err := c.markCallbackCounter(sourceChainID, idx); err != nil {
			return err
		}
	}

	return nil
}

/************* end of index-helper related method area **************/