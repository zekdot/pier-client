package main

import (
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Broker struct {
	contractapi.Contract
}

func (broker *Broker) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// get is business function which will invoke the to,tid,id
func (broker *Broker) Get(ctx contractapi.TransactionContextInterface) (string, error) {
	_, args := ctx.GetStub().GetFunctionAndParameters()

	switch len(args) {
	case 1:
		// args[0]: key
		dataAsBytes, err := ctx.GetStub().GetState(args[0])
		if err != nil {
			return "", err
		}

		return string(dataAsBytes), nil
	default:
		return "", fmt.Errorf("incorrect number of arguments")
	}
}

// get is business function which will invoke the to,tid,id
func (b *Broker) Set(ctx contractapi.TransactionContextInterface) error {
	_, args := ctx.GetStub().GetFunctionAndParameters()

	if len(args) != 2 {
		return fmt.Errorf("incorrect number of arguments")
	}

	err := ctx.GetStub().PutState(args[0], []byte(args[1]))
	if err != nil {
		return err
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(Broker))

	if err != nil {
		fmt.Printf("Error create broker chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting broker chaincode: %s", err.Error())
	}
}
