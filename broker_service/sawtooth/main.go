package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	brokerClient, _ := NewBrokerClient(SAWTOOTH_URL, KEY_PATH)
	service := NewService(brokerClient)
	log.Printf("start listen")
	rpc.Register(service)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":" + RPC_PORT)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	http.Serve(l, nil)
}
