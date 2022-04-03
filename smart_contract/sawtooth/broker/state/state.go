package state

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	"strings"
)

//var Namespace = hexdigest("broker")[:6]
var DataNamespace = "19d832"
var MetaNamespace = "5978b3"

// 直接存储key到另一个字符串的映射
type BrokerState struct {
	context *processor.Context
}

func NewBrokerState(context *processor.Context) *BrokerState {
	return &BrokerState{
		context: context,
	}
}

func (broker *BrokerState) SetMetaData(key string, value string) error {
	address := makeAddress(MetaNamespace, key)
	data := []byte(value)
	// 进行缓存
	//broker.addressCache[address] = data
	fmt.Printf("will save %s to %s\n", data, address)
	// 存储进账本中
	_, err := broker.context.SetState(map[string][] byte {
		address: data,
	})
	return err
}

func (broker *BrokerState) SetData(key string, value string) error {
	address := makeAddress(DataNamespace, key)
	data := []byte(value)
	// 进行缓存
	//broker.addressCache[address] = data
	fmt.Printf("will save %s to %s", value, address)
	// 存储进账本中
	_, err := broker.context.SetState(map[string][] byte {
		address: data,
	})
	return err
}

// split regular data and meta
func makeAddress(prefix string, name string) string {
	//fmt.Printf("get digest of %s\n", (typeValue + name))
	return prefix + "00" + hexdigest(name)[:62]
}

// split regular data and meta
//func makeMetaAddress(prefix string, name string) string {
//	//fmt.Printf("get digest of %s\n", (typeValue + name))
//	return prefix + hexdigest(name)[:64]
//}

func hexdigest(str string) string {
	hash := sha512.New()
	hash.Write([]byte(str))
	hashBytes := hash.Sum(nil)
	return strings.ToLower(hex.EncodeToString(hashBytes))
}

