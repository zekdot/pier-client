package contract

import (
	"broker/state"
	"fmt"
)

const (
	innerMeta = "inner-meta"
	outterMeta = "outter-meta"
	callbackMeta = "callback-meta"
	delimiter = "&"
)

type Broker struct {
	init bool
}

type Event struct {
	Index         uint64 `json:"index"`
	DstChainID    string `json:"dst_chain_id"`
	SrcContractID string `json:"src_contract_id"`
	DstContractID string `json:"dst_contract_id"`
	Func          string `json:"func"`
	Args          string `json:"args"`
	Callback      string `json:"callback"`
}

func NewBroker() *Broker {
	return &Broker{
		init: false,
	}
}

func (broker *Broker) SetData(state *state.BrokerState, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("incorrect number of arguments")
	}
	err := state.SetData(args[0], args[1])
	if err != nil {
		return err
	}
	return nil
}

func (broker *Broker) SetMeta(state *state.BrokerState, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("incorrect number of arguments")
	}
	err := state.SetMetaData(args[0], args[1])
	if err != nil {
		return err
	}
	return nil
}