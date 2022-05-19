package main

import "strconv"

func main() {
	client, err := GetClient()
	if err != nil {
		panic(err)
	}
	for i := 0; i < 100; i ++ {
		client.InterchainGet(PIER_ID, "mychannel&data_swapper", "key" + strconv.FormatInt(int64(i+1), 10))
		//fmt.Println("key" + strconv.FormatInt(int64(i+1), 10))
	}
}

func GetClient() (*RpcClient, error) {
	return NewRpcClient(RPC_URL)
}
