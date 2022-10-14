package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	appchainClient, err := NewAppchainClient(APPCHAIN_ADDRESS)
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
