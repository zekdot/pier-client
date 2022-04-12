package payload

import (
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"encoding/json"
)

type BrokerPayload struct {
	Key string
	Value string
}

func FromBytes(payloadData[] byte) (*BrokerPayload, error) {
	if payloadData == nil {
		return nil, &processor.InvalidTransactionError{Msg: "Must contain payload"}
	}

	payload := &BrokerPayload{}
	if err := json.Unmarshal(payloadData, payload); err != nil {
		return nil, err
	}
	return payload, nil
}
