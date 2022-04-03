package payload

import (
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"encoding/json"
)

type BrokerPayload struct {
	Function string	// the function be called
	Parameter []string	// parameter passed to function
}

func FromBytes(payloadData[] byte) (*BrokerPayload, error) {
	if payloadData == nil {
		return nil, &processor.InvalidTransactionError{Msg: "Must contain payload"}
	}

	payload := &BrokerPayload{}
	if err := json.Unmarshal(payloadData, payload); err != nil {
		return nil, err
	}
	payload.Parameter = payload.Parameter[0:len(payload.Parameter) - 1]
	if len(payload.Function) < 1 {
		return nil, &processor.InvalidTransactionError{Msg: "Function is required"}
	}

	return payload, nil
}
