package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	httpClient := NewHttpClient()
	service := NewService(httpClient)
	log.Printf("start listen")
	rpc.Register(service)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":" + RPC_PORT)
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	http.Serve(l, nil)
}
