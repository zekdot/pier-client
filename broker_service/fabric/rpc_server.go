package main

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"os"
)

var (
	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "client",
		Output: os.Stderr,
		Level:  hclog.Trace,
	})
)

type Service struct {
	broker *BrokerClient
}


type ReqArgs struct {
	FuncName string
	Args []string
}

func NewService(broker *BrokerClient) *Service {
	return &Service{
		broker: broker,
	}
}

// send transaction and don't need result
func (s *Service) SetValue(req *ReqArgs, reply *string) error{

	broker := s.broker
	args := req.Args
	// if this is a outer cross-chain request
	if len(args[0]) > 7 && args[0][0:7] == "out-msg" {
		logger.Debug("s1:save cross-chain request to ledger")
	}
	fmt.Printf("set %s to %s\n", args[0], args[1])
	err := broker.setValue(args[0], args[1])
	return err
}

// query transaction and need result
func (s *Service) GetValue(req *ReqArgs, reply *string) error{
	broker := s.broker
	args := req.Args
	fmt.Printf("get value of %s\n", args[0])
	res, err := broker.getValue(args[0])
	*reply = string(res)
	return err
}