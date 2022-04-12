package handler

import (
	"broker/contract"
	"broker/payload"
	"broker/state"
	"fmt"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/processor_pb2"
)

type BrokerHandler struct {
	broker *contract.Broker
}

func NewHandler(broker *contract.Broker) *BrokerHandler {
	return &BrokerHandler{
		broker: broker,
	}
}

func (handler *BrokerHandler) FamilyName() string {
	return "cross-chain"
}
func(handler *BrokerHandler) FamilyVersions() [] string {
	return []string{"0.1"}
}
func (handler *BrokerHandler) Namespaces()[]string {
	return []string{state.MetaNamespace, state.DataNamespace}
}

func (broker *BrokerHandler) isMetaRequest(key string) bool {
	if len(key) < 4 {
		return false
	}
	var prefix = key[:4]
	return prefix == "inne" || prefix == "outt" || prefix == "call" || prefix == "in-m" || prefix == "out-"
}

func (handler *BrokerHandler) Apply(request *processor_pb2.TpProcessRequest, context *processor.Context) error {
	fmt.Printf("receive %s", string(request.GetPayload()))
	// unmarshal from json bytes
	payload, err := payload.FromBytes(request.GetPayload())
	if err != nil {
		return err
	}
	//fmt.Printf("before context")
	brokerState := state.NewBrokerState(context)
	//fmt.Printf("after context")
	broker := handler.broker
	key := payload.Key
	value := payload.Value
	if handler.isMetaRequest(key) {
		args := []string{key, string(request.GetPayload())}
		return broker.SetMeta(brokerState, args)
	} else {
		args := []string{key, value}
		return broker.SetData(brokerState, args)
	}
}