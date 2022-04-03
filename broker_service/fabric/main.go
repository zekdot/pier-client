package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	brokerClient, err := NewBrokerClient()
	if err != nil {
		log.Fatalf("create broker_client failed: %v", err)
	}
	service := NewService(brokerClient)
	log.Printf("start listen")
	rpc.Register(service)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1212")
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	http.Serve(l, nil)
}