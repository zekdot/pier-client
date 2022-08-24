package main

import (
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"os"
	"strings"
	"sync"
)

var (
	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "performance",
		Output: os.Stdout,
		Level:  hclog.Trace,
	})
)
type Event struct {
	Index         uint64 `json:"index"`
	DstChainID    string `json:"dst_chain_id"`
	SrcContractID string `json:"src_contract_id"`
	DstContractID string `json:"dst_contract_id"`
	Func          string `json:"func"`
	Args          string `json:"args"`
	Callback      string `json:"callback"`
	Argscb        string `json:"argscb"`
	Rollback      string `json:"rollback"`
	Argsrb        string `json:"argsrb"`
	Proof         []byte `json:"proof"`
	Extra         []byte `json:"extra"`
}
type Service struct {
	httpClient *HttpClient
	db *DB
	bundleManager *BundleManager
}

type ReqArgs struct {
	FuncName string
	Args []string
}

func NewService(httpClient *HttpClient) *Service {
	db, err := NewDB(DB_PATH)
	if err != nil {
		panic(err)
	}
	bundleManager:= NewBundleManager()
	return &Service{
		httpClient: httpClient,
		db: db,
		bundleManager: bundleManager,
	}
}

// query transaction and need result
func (s *Service) GetValue(req *ReqArgs, reply *string) error{

	args := req.Args
	consumerPackageId := args[0]
	res, err := s.httpClient.GetValue(consumerPackageId)
	*reply = res
	return err
}

// query transaction and need result
func (s *Service) Init(req *ReqArgs, reply *string) error{
	return s.db.Init()
}

func (s *Service) InterchainGet(req *ReqArgs, reply *string) error {
	args := req.Args

	destChainID := args[0]
	contractId := args[1]
	key := args[2]
	logger.Info("s1:key-" + key + " save cross-chain request to ledger")
	s.bundleManager.AddRequest(destChainID, contractId, key)


	defer logger.Info("s2:key-" + key + " have saved cross-chain request to ledger")

	//return s.db.InterchainGet(destChainID, contractId, key)
	return nil
}



func (s *Service) GetMeta(req *ReqArgs, reply *string) error {
	args := req.Args

	key := args[0]

	value, err := s.db.GetMetaStr(key)
	if err != nil {
		return err
	}
	*reply = value
	return nil
}

func (s *Service) PollingHelper(req *ReqArgs, reply *string) error {
	args := req.Args

	mStr := args[0]

	m := make(map[string]uint64)
	if err := json.Unmarshal([]byte(mStr), &m); err != nil {
		return err
	}
	// just save bundle to db
	destChainID := s.bundleManager.PeekNextPierId()
	if destChainID != nil {
		//logger.Debug("检测到发往" + *destChainID + "的交易")
		tx, err := s.bundleManager.GetFirstBundle()

		destChainIds := strings.Split(*destChainID, "#")
		//
		//temp, _ := json.Marshal(tx)
		//logger.Debug("内容为" + string(temp))

		if err != nil {
			return err
		}
		if err := s.db.SaveInterchainReq(destChainIds[0], tx); err != nil {
			return err
		}
	}
	evs, err := s.db.PollingEvents(m)
	if err != nil {
		return err
	}
	evsStr, err := json.Marshal(evs)
	*reply = string(evsStr)
	return err
}

func (s *Service) GetInMessageStrByKey(req *ReqArgs, reply *string) error {
	args := req.Args
	key := args[0]
	var err error
	*reply, err = s.db.GetInMessageStrByKey(key)
	return err
}

func (s *Service) GetOutMessageStrByKey(req *ReqArgs, reply *string) error {
	args := req.Args
	key := args[0]
	var err error
	*reply, err = s.db.GetOutMessageStrByKey(key)
	return err
}

func MultiRead(threadNum int, keys []string, client *HttpClient) (string, error) {
	//threadNum := 10
	ch := make(chan string, len(keys))
	//done := make(chan bool, 5)
	res := make([][]string, 0)
	var wg sync.WaitGroup
	if len(keys) < threadNum {
		threadNum = len(keys)
	}
	wg.Add(threadNum)
	for j := 0; j < threadNum; j ++ {
		go func(index uint64) {
			var key string
			for true {
				select {
				case key =<- ch:
					if key == "done" {
						//fmt.Println("进程" + strconv.FormatUint(index, 10) + "结束")
						wg.Done()
						return
					}
					logger.Info("s6:key-" + key + " try to get value from sawtooth")
					valueBytes, _ := client.GetValue(key)
					logger.Info("s7:key-" + key + " get value " + string(valueBytes) + " from sawtooth successfully")
					res = append(res, []string{key, string(valueBytes)})
					//fmt.Println("处理" + key)
					//time.Sleep(1 * time.Millisecond)
				}
			}

		}(uint64(j))
	}
	for _, key := range keys {
		ch <- key
	}
	for j := 0; j < threadNum; j ++ {
		ch <- "done"
	}
	wg.Wait()
	resStr, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(resStr), nil
}

func (s *Service) InvokeInterchainHelper(req *ReqArgs, reply *string) error {
	args := req.Args
	sourceChainID := args[0]
	sequenceNum := args[1]
	targetCID := args[2]
	isReq := args[3]
	funcName := args[4]
	// no matter keys or key-values all in this parameter
	arg := args[5]
	//var value string
	//if len(args) > 6 {
	//	value = args[6]
	//} else {
	//	value = ""
	//}
	var err error
	value := ""
	if funcName == "bundleRequest" {
		keys := make([]string, 0)
		if err = json.Unmarshal([]byte(arg), &keys); err != nil {
			return err
		}
		value, err = MultiRead(10, keys, s.httpClient)
		*reply = value
	} else if funcName == "bundleResponse" {
		// I think sawtooth side won't
		//kvpairs := make([][]string, 0)
		//if err = json.Unmarshal([]byte(arg), &kvpairs); err != nil {
		//	return err
		//}
		//for _, kv := range kvpairs {
		//	err = s.broker.setValue(kv[0], kv[1])
		//}
	}
	s.db.InvokeInterchainHelper(sourceChainID, sequenceNum, targetCID, isReq, funcName, value)
	return err
}

func (s *Service) UpdateIndexHelper(req *ReqArgs, reply *string) error {
	args := req.Args
	sourceChainID := args[0]
	sequenceNum := args[1]
	isReq := args[2]
	return s.db.UpdateIndex(sourceChainID, sequenceNum, isReq)
}
