package main

import (
	bytes2 "bytes"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gogo/protobuf/proto"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/batch_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/protobuf/transaction_pb2"
	"github.com/hyperledger/sawtooth-sdk-go/signing"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BrokerClient struct {
	url string
	signer *signing.Signer
}

func NewBrokerClient(url string, keyfile string) (*BrokerClient, error) {
	//fmt.Println(url)
	//fmt.Println(keyfile)
	var privateKey signing.PrivateKey
	if keyfile != "" {
		// Read private key file
		privateKeyStr, err := ioutil.ReadFile(keyfile)
		// fmt.Println(privateKeyStr)
		if err != nil {
			return &BrokerClient{},
				errors.New(fmt.Sprintf("Failed to read private key: %v", err))
		}
		// Get private key object
		privateKey = signing.NewSecp256k1PrivateKey(privateKeyStr)
	} else {
		privateKey = signing.NewSecp256k1Context().NewRandomPrivateKey()
	}
	cryptoFactory := signing.NewCryptoFactory(signing.NewSecp256k1Context())
	signer := cryptoFactory.NewSigner(privateKey)
	//fmt.Println(signer)
	return &BrokerClient{url, signer}, nil
}

func (broker *BrokerClient) isMetaRequest(key string) bool {
	if len(key) < 4 {
		return false
	}
	var prefix = key[:4]
	return prefix == "inne" || prefix == "outt" || prefix == "call" || prefix == "in-m" || prefix == "out-"
}

// only need to fetch value according to the key
func (broker *BrokerClient)getValue(key string) ([]byte, error) {
	var address string
	var isMeta = broker.isMetaRequest(key)
	if isMeta {
		address = broker.getAddress(META_NAMESPACE, key)
	} else {
		address = broker.getAddress(DATA_NAMESPACE, key)
	}
	apiSuffix := fmt.Sprintf("%s/%s", STATE_API, address)
	//log.Printf("apiSuffix is %s\n", apiSuffix)
	//fmt.Printf()

	rawData, err := broker.sendRequest(apiSuffix, []byte{}, "", key)
	//log.Printf("Get raw data: %s\n", rawData)
	if err != nil {
		return nil, err
	}
	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(rawData), &responseMap)
	if err != nil {
		return nil, errors.New(fmt.Sprint("Error reading response: %v", err))
	}
	data, _ := responseMap["data"].(string)
	//log.Printf("After base64 decode: %s\n", data)


	fishStr, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, err
	}
	//log.Printf("After base64 decode: %s\n", string(fishStr))
	if isMeta {
		resMap := make(map[string]string)
		err := json.Unmarshal(fishStr, &resMap)
		if err != nil {
			return nil, err
		}
		//log.Println(resMap)
		return []byte(resMap["value"]), nil
	}
	return fishStr, nil
}

// only need to set value according to the key and value
func (broker *BrokerClient)setValue(key string, value string) error {
	var err error
	if broker.isMetaRequest(key) {
		_, err = broker.sendTransaction("setMeta", key, value, 0)
	} else {
		// in fact, in our situation, there won't be setData be called
		_, err = broker.sendTransaction("setData", key, value, 0)
	}
	if err != nil {
		return err
	}
	return nil
}

func (broker *BrokerClient) sendRequest(
	apiSuffix string,
	data []byte,
	contentType string,
	name string) (string, error) {

	// Construct URL
	var url string
	url = fmt.Sprintf("%s/%s", SAWTOOTH_URL, apiSuffix)
	//fmt.Println("url" + url)
	// Send request to validator URL
	var response *http.Response
	var err error
	// fmt.Println("data")
	// fmt.Println(string(data))

	if len(data) > 0 {
		response, err = http.Post(url, contentType, bytes2.NewBuffer(data))
		//fmt.Println("response")
		//fmt.Println(response)
	} else {
		response, err = http.Get(url)
	}
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Failed to connect to REST API: %v", err))
	}
	if response.StatusCode == 404 {
		//logger.Debug(fmt.Sprintf("%v", response))
		return "", errors.New(fmt.Sprintf("No such key: %s", name))
	} else if response.StatusCode >= 400 {
		return "", errors.New(
			fmt.Sprintf("Error %d: %s", response.StatusCode, response.Status))
	}
	defer response.Body.Close()
	reponseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	return string(reponseBody), nil
}

func (broker *BrokerClient) getStatus(
	batchId string, wait uint) (string, error) {

	// API to call
	apiSuffix := fmt.Sprintf("%s?id=%s&wait=%d",
		BATCH_STATUS_API, batchId, wait)
	response, err := broker.sendRequest(apiSuffix, []byte{}, "", "")
	if err != nil {
		return "", err
	}

	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(response), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error reading response: %v", err))
	}
	entry :=
		responseMap["data"].([]interface{})[0].(map[interface{}]interface{})
	return fmt.Sprint(entry["status"]), nil
}

func (broker *BrokerClient) sendTransaction(
	function string, key string, value string, wait uint) (string, error) {
	rand.Seed(time.Now().Unix())

	payloadData := make(map[string]interface{})
	payloadData["key"] = key
	payloadData["value"] = value
	// payloadData["Function"] = function
	//	args := make([]string, 0)
	//args = append(args, key, value, strconv.Itoa(rand.Int()))
	//payloadData["Parameter"] = args
	payload, err := json.Marshal(payloadData)
	//fmt.Println("payload", string(payload))
	if err != nil {
		return "", err
	}
	// construct the address
	var address string
	if function == "setMeta" {
		// address = broker.getAddress(META_NAMESPACE, key)
		address = META_NAMESPACE
	} else if function == "setData" {
		address = DATA_NAMESPACE
	}
	//log.Printf("save to address hash %v\n", address)
	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  broker.signer.GetPublicKey().AsHex(),
		FamilyName:       FAMILY_NAME,
		FamilyVersion:    FAMILY_VERSION,
		Dependencies:     []string{}, // empty dependency list
		Nonce:            strconv.Itoa(rand.Int()),
		BatcherPublicKey: broker.signer.GetPublicKey().AsHex(),
		Inputs:           []string{address},
		Outputs:          []string{address},
		PayloadSha512:    Sha512HashValue(string(payload)),
	}
	transactionHeader, err := proto.Marshal(&rawTransactionHeader)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize transaction header: %v", err))
	}

	// Signature of TransactionHeader
	transactionHeaderSignature := hex.EncodeToString(
		broker.signer.Sign(transactionHeader))

	// Construct Transaction
	transaction := transaction_pb2.Transaction{
		Header:          transactionHeader,
		HeaderSignature: transactionHeaderSignature,
		Payload:         []byte(payload),
	}

	// Get BatchList
	rawBatchList, err := broker.createBatchList(
		[]*transaction_pb2.Transaction{&transaction})
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to construct batch list: %v", err))
	}
	batchId := rawBatchList.Batches[0].HeaderSignature
	batchList, err := proto.Marshal(&rawBatchList)
	if err != nil {
		return "", errors.New(
			fmt.Sprintf("Unable to serialize batch list: %v", err))
	}

	if wait > 0 {
		waitTime := uint(0)
		startTime := time.Now()
		response, err := broker.sendRequest(
			BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, key)
		if err != nil {
			return "", err
		}
		for waitTime < wait {
			status, err := broker.getStatus(batchId, wait-waitTime)
			if err != nil {
				return "", err
			}
			waitTime = uint(time.Now().Sub(startTime))
			if status != "PENDING" {
				return response, nil
			}
		}
		return response, nil
	}

	return broker.sendRequest(
		BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, key)
}

func (broker *BrokerClient) createBatchList(
	transactions []*transaction_pb2.Transaction) (batch_pb2.BatchList, error) {

	// Get list of TransactionHeader signatures
	transactionSignatures := []string{}
	for _, transaction := range transactions {
		transactionSignatures =
			append(transactionSignatures, transaction.HeaderSignature)
	}

	// Construct BatchHeader
	rawBatchHeader := batch_pb2.BatchHeader{
		SignerPublicKey: broker.signer.GetPublicKey().AsHex(),
		TransactionIds:  transactionSignatures,
	}
	batchHeader, err := proto.Marshal(&rawBatchHeader)
	if err != nil {
		return batch_pb2.BatchList{}, errors.New(
			fmt.Sprintf("Unable to serialize batch header: %v", err))
	}

	// Signature of BatchHeader
	batchHeaderSignature := hex.EncodeToString(
		broker.signer.Sign(batchHeader))

	// Construct Batch
	batch := batch_pb2.Batch{
		Header:          batchHeader,
		Transactions:    transactions,
		HeaderSignature: batchHeaderSignature,
	}

	// Construct BatchList
	return batch_pb2.BatchList{
		Batches: []*batch_pb2.Batch{&batch},
	}, nil
}

func Sha512HashValue(value string) string {
	hashHandler := sha512.New()
	hashHandler.Write([]byte(value))
	return strings.ToLower(hex.EncodeToString(hashHandler.Sum(nil)))
}

func (broker *BrokerClient) getAddress(prefix, name string) string {
	//prefix := broker.getPrefix()
	nameAddress := "00" + Sha512HashValue(name)[:FAMILY_VERB_ADDRESS_LENGTH]
	return prefix + nameAddress
}
