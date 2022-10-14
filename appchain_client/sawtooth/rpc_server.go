package main

import (
	"github.com/hashicorp/go-hclog"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)

var (
	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "performance",
		Output: os.Stdout,
		Level:  hclog.Trace,
	})
)

type Service struct {
	appchainClient *AppchainClient
}

type ReqArgs struct {
	Args []string
}

func NewService(appchainClient *AppchainClient) *Service {
	return &Service{
		appchainClient: appchainClient,
	}
}

// send transaction and don't need result
func (s *Service) SetValue(req *ReqArgs, reply *string) error{
	args := req.Args
	err := (*s.appchainClient).SetValue(args[0], args[1])
	return err
}

// query transaction and need result
func (s *Service) GetValue(req *ReqArgs, reply *string) error {
	args := req.Args
	consumerPackageId := args[0]
	res, err := (*s.appchainClient).GetValue(consumerPackageId)
	*reply = res
	return err
}

func main() {
	appchainClient, err := NewAppchainClient(SAWTOOTH_URL, KEY_PATH)
	if err != nil {
		log.Fatal("listen error: ", err)
	}
	service := NewService(appchainClient)
	log.Printf("start listen")
	rpc.Register(service)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":" + RPC_PORT)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	http.Serve(l, nil)
}