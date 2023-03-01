package main

import (
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"log"
	"os"
	"path/filepath"
)

type FabricClient struct {
	contract *gateway.Contract
}

func NewFabricClient() (*FabricClient, error) {
	log.Println("============ application-golang starts ============")
	//reqConfig := &Fabric{
	//	Addr:        Addr,
	//	OrganizationsPath:        OrganizationsPath,
	//	Username:    Username,
	//	CCID:        CCID,
	//	ChannelId:   ChannelId,
	//}

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		return nil, err
	}
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}

	if !wallet.Exists(Username) {
		err = populateWallet(wallet, OrganizationsPath, Username)
		if err != nil {
			log.Fatalf("Failed to populate wallet contents: %v", err)
		}
	}

	ccpPath := filepath.Join(
		OrganizationsPath,
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)
	//log.Println(ccpPath)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, Username),
	)
	if err != nil {
		//log.Fatalf("Failed to connect to gateway: %v", err)
		return nil, err
	}
	defer gw.Close()

	network, err := gw.GetNetwork(ChannelId)
	if err != nil {
		//log.Fatalf("Failed to get network: %v", err)
		return nil, err
	}

	contract := network.GetContract(CCID)
	//s.contract = contract
	return &FabricClient{
		contract: contract,
	}, nil
}

func (client *FabricClient) getValue(key string) ([]byte, error) {
	var result []byte
	var err error
	//if len(req.Args) == 0 {
	//	result, err = s.contract.EvaluateTransaction(req.FuncName)
	//} else {
	result, err = client.contract.EvaluateTransaction("Get", key)
	//}
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (client *FabricClient) setValue(key string, value string) error {
	_, err := client.contract.SubmitTransaction("Set", key, value)

	if err != nil {
		return err
	}
	return nil
}

//func populateWallet(wallet *gateway.Wallet, organizationsPath string, username string) error {
//	credPath := filepath.Join(
//		organizationsPath,
//		"peerOrganizations",
//		"org1.example.com",
//		"users",
//		"User1@org1.example.com",
//		"msp",
//	)
//
//	certPath := filepath.Join(credPath, "signcerts", "User1@org1.example.com-cert.pem")
//	// read the certificate pem
//	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
//	if err != nil {
//		return err
//	}
//
//	keyDir := filepath.Join(credPath, "keystore")
//	// there's a single file in this dir containing the private key
//	files, err := ioutil.ReadDir(keyDir)
//	if err != nil {
//		return err
//	}
//	if len(files) != 1 {
//		return fmt.Errorf("keystore folder should have contain one file")
//	}
//	keyPath := filepath.Join(keyDir, files[0].Name())
//	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
//	if err != nil {
//		return err
//	}
//
//	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))
//
//	return wallet.Put(username, identity)
//}