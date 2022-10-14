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

type SawtoothClient struct {
	url string
	signer *signing.Signer
}

func NewAppchainClient(url string, keyfile string) (*AppchainClient, error) {
	//fmt.Println(url)
	//fmt.Println(keyfile)
	var privateKey signing.PrivateKey
	if keyfile != "" {
		// Read private key file
		privateKeyStr, err := ioutil.ReadFile(keyfile)
		// fmt.Println(privateKeyStr)
		if err != nil {
			return nil,
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
	var sawtoothClient AppchainClient = &SawtoothClient{
		url,
		signer,
	}
	return &sawtoothClient, nil
}

// only need to fetch value according to the key
func (client *SawtoothClient)GetValue(key string) (string, error) {
	var address string
	address = client.getAddress(DATA_NAMESPACE, key)

	apiSuffix := fmt.Sprintf("%s/%s", STATE_API, address)
	//log.Printf("apiSuffix is %s\n", apiSuffix)
	//fmt.Printf()

	rawData, err := client.sendRequest(apiSuffix, []byte{}, "", key)
	//log.Printf("Get raw data: %s\n", rawData)
	if err != nil {
		return "", err
	}
	responseMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal([]byte(rawData), &responseMap)
	if err != nil {
		return "", errors.New(fmt.Sprint("Error reading response: %v", err))
	}
	data, _ := responseMap["data"].(string)
	//log.Printf("After base64 decode: %s\n", data)


	fishStr, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return "", err
	}
	//log.Printf("After base64 decode: %s\n", string(fishStr))
	return string(fishStr), nil
}

// only need to set value according to the key and value
func (client *SawtoothClient)SetValue(key string, value string) error {
	var err error

	// in fact, in our situation, there won't be setData be called
	_, err = client.sendTransaction("setData", key, value, 0)

	if err != nil {
		return err
	}
	return nil
}

func (client *SawtoothClient) sendRequest(
	apiSuffix string,
	data []byte,
	contentType string,
	name string) (string, error) {

	// Construct URL
	var url string
	url = fmt.Sprintf("%s/%s", SAWTOOTH_URL, apiSuffix)
	// Send request to validator URL
	var response *http.Response
	var err error

	if len(data) > 0 {
		response, err = http.Post(url, contentType, bytes2.NewBuffer(data))
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

func (client *SawtoothClient) getStatus(
	batchId string, wait uint) (string, error) {

	// API to call
	apiSuffix := fmt.Sprintf("%s?id=%s&wait=%d",
		BATCH_STATUS_API, batchId, wait)
	response, err := client.sendRequest(apiSuffix, []byte{}, "", "")
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

func (client *SawtoothClient) sendTransaction(
	function string, key string, value string, wait uint) (string, error) {
	rand.Seed(time.Now().Unix())

	payloadData := make(map[string]interface{})
	payloadData["key"] = key
	payloadData["value"] = value

	payload, err := json.Marshal(payloadData)
	//fmt.Println("payload", string(payload))
	if err != nil {
		return "", err
	}
	// construct the address
	var address string
	address = DATA_NAMESPACE

	//log.Printf("save to address hash %v\n", address)
	// Construct TransactionHeader
	rawTransactionHeader := transaction_pb2.TransactionHeader{
		SignerPublicKey:  client.signer.GetPublicKey().AsHex(),
		FamilyName:       FAMILY_NAME,
		FamilyVersion:    FAMILY_VERSION,
		Dependencies:     []string{}, // empty dependency list
		Nonce:            strconv.Itoa(rand.Int()),
		BatcherPublicKey: client.signer.GetPublicKey().AsHex(),
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
		client.signer.Sign(transactionHeader))

	// Construct Transaction
	transaction := transaction_pb2.Transaction{
		Header:          transactionHeader,
		HeaderSignature: transactionHeaderSignature,
		Payload:         payload,
	}

	// Get BatchList
	rawBatchList, err := client.createBatchList(
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
		response, err := client.sendRequest(
			BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, key)
		if err != nil {
			return "", err
		}
		for waitTime < wait {
			status, err := client.getStatus(batchId, wait-waitTime)
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

	return client.sendRequest(
		BATCH_SUBMIT_API, batchList, CONTENT_TYPE_OCTET_STREAM, key)
}

func (client *SawtoothClient) createBatchList(
	transactions []*transaction_pb2.Transaction) (batch_pb2.BatchList, error) {

	// Get list of TransactionHeader signatures
	transactionSignatures := []string{}
	for _, transaction := range transactions {
		transactionSignatures =
			append(transactionSignatures, transaction.HeaderSignature)
	}

	// Construct BatchHeader
	rawBatchHeader := batch_pb2.BatchHeader{
		SignerPublicKey: client.signer.GetPublicKey().AsHex(),
		TransactionIds:  transactionSignatures,
	}
	batchHeader, err := proto.Marshal(&rawBatchHeader)
	if err != nil {
		return batch_pb2.BatchList{}, errors.New(
			fmt.Sprintf("Unable to serialize batch header: %v", err))
	}

	// Signature of BatchHeader
	batchHeaderSignature := hex.EncodeToString(
		client.signer.Sign(batchHeader))

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

func (client *SawtoothClient) getAddress(prefix, name string) string {
	//prefix := broker.getPrefix()
	nameAddress := "00" + Sha512HashValue(name)[:FAMILY_VERB_ADDRESS_LENGTH]
	return prefix + nameAddress
}
