package main

import (
	"encoding/json"
	"strings"
)

type BundleManager struct {
	// one queue to save all pierId
	queue Queue
	// one map to save pierId to all transactions
	address2Trans map[string] []string
}

func NewBundleManager() *BundleManager {

	//q := NewArrayQueue()
	var q Queue = NewArrayQueue()
	m := make(map[string][]string)
	return &BundleManager{
		queue: q,
		address2Trans: m,
	}
}

func address(pierId, contractId string) string {
	return pierId + "#" + contractId
}

func (bundleManager *BundleManager)AddRequest(pierId, contractId string, key string) error {
	// if this request already in queue, just add it to map
	addValue := address(pierId, contractId)
	if _, ok := bundleManager.address2Trans[addValue]; !ok {
		bundleManager.queue.Enqueue(addValue)
		keys := make([]string, 0)
		keys = append(keys, key)
		bundleManager.address2Trans[addValue] = keys
		return nil
	}
	bundleManager.address2Trans[addValue] = append(bundleManager.address2Trans[addValue], key)
	return nil
}

func (bundleManager *BundleManager)GetFirstBundle() (*Event, error){
	address := bundleManager.queue.Dequeue()
	if address == nil {
		return nil, nil
	}
	keys, err := json.Marshal(bundleManager.address2Trans[*address])
	if err != nil {
		return nil, err
	}
	addressParts := strings.Split(*address, "#")
	pierId := addressParts[0]
	contractId := addressParts[1]
	ev := &Event{
		Index: 0,
		DstChainID: pierId,
		SrcContractID: contractId,
		DstContractID: contractId,
		Func: "bundleRequest",
		Args: string(keys),
		Callback: "bundleResponse",
		// In fact here we don't need callback parameter anymore because it is contained in value
		//Argscb: string(keys),
		Argscb: "",
		Rollback: "",
		Argsrb: "",
	}
	delete(bundleManager.address2Trans, *address)
	return ev, nil
}

func (bundleManager *BundleManager)PeekNextPierId() *string {
	return bundleManager.queue.Peek()
}