package main

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
)

func main() {
	brokerClient, _ := NewBrokerClient(SAWTOOTH_URL, "/home/hzh/.sawtooth/keys/hzh.priv")
	service := NewService(brokerClient)
	log.Printf("start listen")
	rpc.Register(service)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1212")
	if e != nil {
		log.Fatal("listen error: ", e)
	}
	http.Serve(l, nil)


	//err := brokerClient.setValue("k2", "v1")
	//if err != nil {
	//	fmt.Errorf(err.Error())
	//}
	//res, err := brokerClient.getValue("k2")
	//if err != nil {
	//	fmt.Errorf(err.Error())
	//}
	//fmt.Println(string(res))
	//value, err := brokerClient.Get("k1")
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//fmt.Println(value)
	//outMetaStr, err := brokerClient.getMeta("getOuterMeta")
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//fmt.Println(outMetaStr)
	//eventStr, err := brokerClient.pollingEvent("{}")
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//fmt.Println(eventStr)
}
