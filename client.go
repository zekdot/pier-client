package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/strategy"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/meshplus/bitxhub-model/pb"
	"github.com/meshplus/pier/pkg/plugins"
)

var (
	logger = hclog.New(&hclog.LoggerOptions{
		Name:   "client",
		Output: os.Stderr,
		Level:  hclog.Trace,
	})
)

var _ plugins.Client = (*Client)(nil)
//
//const (
//	GetInnerMetaMethod      = "getInnerMeta"    // get last index of each source chain executing tx
//	GetOutMetaMethod        = "getOuterMeta"    // get last index of each receiving chain crosschain event
//	GetCallbackMetaMethod   = "getCallbackMeta" // get last index of each receiving chain callback tx
//	GetInMessageMethod      = "getInMessage"
//	GetOutMessageMethod     = "getOutMessage"
//	PollingEventMethod      = "pollingEvent"
//	InvokeInterchainMethod  = "invokeInterchain"
//	InvokeIndexUpdateMethod = "invokeIndexUpdate"
//	FabricType              = "fabric"
//)

type ContractMeta struct {
	EventFilter string `json:"event_filter"`
	Username    string `json:"username"`
	CCID        string `json:"ccid"`
	ChannelID   string `json:"channel_id"`
	ORG         string `json:"org"`
}

type Client struct {
	meta     *ContractMeta
	//consumer *Consumer
	eventC   chan *pb.IBTP
	pierId   string
	name     string
	outMeta  map[string]uint64
	ticker   *time.Ticker
	done     chan bool
	client   *RpcClient
}

type CallFunc struct {
	Func string   `json:"func"`
	Args [][]byte `json:"args"`
}

func (c *Client) Initialize(configPath, pierId string, extra []byte) error {
	eventC := make(chan *pb.IBTP)
	//fabricConfig, err := UnmarshalConfig(configPath)
	//if err != nil {
	//	return fmt.Errorf("unmarshal config for plugin :%w", err)
	//}
	//
	//contractmeta := &ContractMeta{
	//	EventFilter: fabricConfig.EventFilter,
	//	Username:    fabricConfig.Username,
	//	CCID:        fabricConfig.CCID,
	//	ChannelID:   fabricConfig.ChannelId,
	//	ORG:         fabricConfig.Org,
	//}

	m := make(map[string]uint64)
	if err := json.Unmarshal(extra, &m); err != nil {
		return fmt.Errorf("unmarshal extra for plugin :%w", err)
	}
	if m == nil {
		m = make(map[string]uint64)
	}

	//mgh, err := newFabricHandler(contractmeta.EventFilter, eventC, pierId)
	//if err != nil {
	//	return err
	//}

	done := make(chan bool)
	//csm, err := NewConsumer(configPath, contractmeta, mgh, done)
	//if err != nil {
	//	return err
	//}
	rpcClient, err := NewRpcClient(RPC_URL)
	if err != nil {
		logger.Error("dialing: ", err)
	}
	c.client = rpcClient
	//c.consumer = csm
	c.eventC = eventC
	//c.meta = contractmeta
	c.pierId = pierId
	c.name = APPCHAIN_TYPE
	c.outMeta = m
	c.ticker = time.NewTicker(2 * time.Second)
	c.done = done

	return nil
}

func (c *Client) Start() error {
	logger.Info("Fabric consumer started")
	go c.polling()
	return nil
	//return c.consumer.Start()
}

// polling event from broker
func (c *Client) polling() {
	for {
		select {
		case <-c.ticker.C:
			evs, err := c.client.Polling(c.outMeta)
			if err != nil {
				return
			}
			for _, ev := range evs {
				ev.Proof = []byte("success")
				evStr, _ := json.Marshal(ev)
				logger.Info("in this polling, event is " + string(evStr))
				c.eventC <- ev.Convert2IBTP(c.pierId, pb.IBTP_INTERCHAIN)
				if c.outMeta == nil {
					c.outMeta = make(map[string]uint64)
				}
				c.outMeta[ev.DstChainID]++
			}
		case <-c.done:
			logger.Info("Stop long polling")
			return
		}
	}
}

func (c *Client) Stop() error {
	c.ticker.Stop()
	c.done <- true
	//return c.consumer.Shutdown()
	return nil
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) Type() string {
	return APPCHAIN_TYPE
}

func (c *Client) GetIBTP() chan *pb.IBTP {
	return c.eventC
}

func (c *Client) SubmitIBTP(ibtp *pb.IBTP) (*pb.SubmitIBTPResponse, error) {
	pd := &pb.Payload{}
	ret := &pb.SubmitIBTPResponse{}
	if err := pd.Unmarshal(ibtp.Payload); err != nil {
		return ret, fmt.Errorf("ibtp payload unmarshal: %w", err)
	}
	content := &pb.Content{}
	if err := content.Unmarshal(pd.Content); err != nil {
		return ret, fmt.Errorf("ibtp content unmarshal: %w", err)
	}

	if ibtp.Category() == pb.IBTP_UNKNOWN {
		return nil, fmt.Errorf("invalid ibtp category")
	}

	logger.Info("submit ibtp", "id", ibtp.ID(), "contract", content.DstContractId, "func", content.Func)
	for i, arg := range content.Args {
		logger.Info("arg", strconv.Itoa(i), string(arg))
	}

	if ibtp.Category() == pb.IBTP_RESPONSE && content.Func == "" {
		logger.Info("InvokeIndexUpdate", "ibtp", ibtp.ID())
		_, resp, err := c.InvokeIndexUpdate(ibtp.From, ibtp.Index, ibtp.Category())
		if err != nil {
			return nil, err
		}
		ret.Status = resp.OK
		ret.Message = resp.Message

		return ret, nil
	}

	var result [][]byte
	//var chResp *channel.Response
	callFunc := CallFunc{
		Func: content.Func,
		Args: content.Args,
	}
	bizData, err := json.Marshal(callFunc)
	if err != nil {
		ret.Status = false
		ret.Message = fmt.Sprintf("marshal ibtp %s func %s and args: %s", ibtp.ID(), callFunc.Func, err.Error())

		_, _, err := c.InvokeIndexUpdate(ibtp.From, ibtp.Index, ibtp.Category())
		if err != nil {
			return nil, err
		}
		//chResp = res
	} else {
		_, resp, err := c.InvokeInterchain(ibtp.From, ibtp.Index, content.DstContractId, ibtp.Category(), bizData)
		if err != nil {
			return nil, fmt.Errorf("invoke interchain for ibtp %s to call %s: %w", ibtp.ID(), content.Func, err)
		}

		ret.Status = resp.OK
		ret.Message = resp.Message

		// if there is callback function, parse returned value
		result = toChaincodeArgs(strings.Split(string(resp.Data), ",")...)
	}

	// If is response IBTP, then simply return
	if ibtp.Category() == pb.IBTP_RESPONSE {
		return ret, nil
	}

	//proof, err := c.getProof(*chResp)
	proof := []byte("success")
	if err != nil {
		return ret, err
	}


	//logger.Info("result is " + string(result))
	ret.Result, err = c.generateCallback(ibtp, result, proof, ret.Status)
	if err != nil {
		return nil, err
	}

	tmp, err := json.Marshal(ret)
	logger.Info("final return of submit IBTP is " + string(tmp))
	return ret, nil
}

func (c *Client) InvokeInterchain(from string, index uint64, destAddr string, category pb.IBTP_Category, bizCallData []byte) (*channel.Response, *Response, error) {
	req := true
	if category == pb.IBTP_RESPONSE {
		req = false
	}
	//args := util.ToChaincodeArgs(from, strconv.FormatUint(index, 10), destAddr, req)
	//args = append(args, bizCallData)
	//request := channel.Request{
	//	ChaincodeID: c.meta.CCID,
	//	Fcn:         InvokeInterchainMethod,
	//	Args:        args,
	//}

	// retry executing
	var res string
	var err error
	if err := retry.Retry(func(attempt uint) error {
		res, err = c.client.InvokeInterchain(from, strconv.FormatUint(index, 10), destAddr, req, bizCallData)
		//res, err = c.consumer.ChannelClient.Execute(request)
		logger.Info("res is " + res)
		if err != nil {
			if strings.Contains(err.Error(), "Chaincode status Code: (500)") {
				//res.ChaincodeStatus = shim.ERROR
				logger.Error("execute request failed", "err", err.Error())
				return nil
			}
			return fmt.Errorf("execute request: %w", err)
		}

		return nil
	}, strategy.Wait(2*time.Second)); err != nil {
		logger.Error("Can't send rollback ibtp back to bitxhub", "err", err.Error())
	}

	if err != nil {
		return nil, nil, err
	}

	//logger.Info("response", "cc status", strconv.Itoa(int(res.ChaincodeStatus)), "payload", string(res.Payload))
	response := &Response{}
	//if err := json.Unmarshal(res.Payload, response); err != nil {
	//	return nil, nil, err
	//}
	response.Data = []byte(res)
	response.OK = true
	return nil, response, nil
}

func (c *Client) GetOutMessage(to string, idx uint64) (*pb.IBTP, error) {
	ret, err := c.client.GetOutMessage(to, idx)
	if err != nil {
		return nil, err
	}
	return ret.Convert2IBTP(c.pierId, pb.IBTP_INTERCHAIN), nil
}

func (c *Client) GetInMessage(from string, index uint64) ([][]byte, error) {
	return c.client.GetInMessage(from, index)
}

func (c *Client) GetInMeta() (map[string]uint64, error) {
	return c.client.GetInnerMeta()
}

func (c *Client) GetOutMeta() (map[string]uint64, error) {
	return c.client.GetOuterMeta()
}

func (c Client) GetCallbackMeta() (map[string]uint64, error) {
	return c.client.GetCallbackMeta()
}

func (c *Client) CommitCallback(ibtp *pb.IBTP) error {
	return nil
}

// @ibtp is the original ibtp merged from this appchain
func (c *Client) RollbackIBTP(ibtp *pb.IBTP, isSrcChain bool) (*pb.RollbackIBTPResponse, error) {
	ret := &pb.RollbackIBTPResponse{}
	pd := &pb.Payload{}
	if err := pd.Unmarshal(ibtp.Payload); err != nil {
		return nil, fmt.Errorf("ibtp payload unmarshal: %w", err)
	}
	content := &pb.Content{}
	if err := content.Unmarshal(pd.Content); err != nil {
		return ret, fmt.Errorf("ibtp content unmarshal: %w", err)
	}

	// only support rollback for interchainCharge
	if content.Func != "interchainCharge" {
		return nil, nil
	}

	callFunc := CallFunc{
		Func: content.Rollback,
		Args: content.ArgsRb,
	}
	bizData, err := json.Marshal(callFunc)
	if err != nil {
		return ret, err
	}

	// pb.IBTP_RESPONSE indicates it is to update callback counter
	_, resp, err := c.InvokeInterchain(ibtp.To, ibtp.Index, content.SrcContractId, pb.IBTP_RESPONSE, bizData)
	if err != nil {
		return nil, fmt.Errorf("invoke interchain for ibtp %s to call %s: %w", ibtp.ID(), content.Rollback, err)
	}

	ret.Status = resp.OK
	ret.Message = resp.Message

	return ret, nil
}

func (c *Client) IncreaseInMeta(original *pb.IBTP) (*pb.IBTP, error) {
	_, _, err := c.InvokeIndexUpdate(original.From, original.Index, original.Category())
	if err != nil {
		logger.Error("update in meta", "ibtp_id", original.ID(), "error", err.Error())
		return nil, err
	}
	//proof, err := c.getProof(*response)
	proof := []byte("success")
	if err != nil {
		return nil, err
	}
	ibtp, err := c.generateCallback(original, nil, proof, false)
	if err != nil {
		return nil, err
	}
	return ibtp, nil
}

func (c *Client) GetReceipt(ibtp *pb.IBTP) (*pb.IBTP, error) {
	result, err := c.GetInMessage(ibtp.From, ibtp.Index)
	if err != nil {
		return nil, err
	}

	status, err := strconv.ParseBool(string(result[0]))
	if err != nil {
		return nil, err
	}
	return c.generateCallback(ibtp, result[1:], nil, status)
}

func (c Client) InvokeIndexUpdate(from string, index uint64, category pb.IBTP_Category) (*channel.Response, *Response, error) {
	req := true
	if category == pb.IBTP_RESPONSE {
		req = false
	}
	err := c.client.InvokeIndexUpdate(from, strconv.FormatUint(index, 10), req)
	if err != nil {
		return nil, nil, err
	}
	response := &Response{}
	//if err := json.Unmarshal(res.Payload, response); err != nil {
	//	return nil, nil, err
	//}
	response.OK = true

	return nil, response, nil
}

func (c *Client) unpackIBTP(response *channel.Response, ibtpType pb.IBTP_Type) (*pb.IBTP, error) {
	ret := &Event{}
	if err := json.Unmarshal(response.Payload, ret); err != nil {
		return nil, err
	}

	return ret.Convert2IBTP(c.pierId, ibtpType), nil
}

func (c *Client) unpackMap(response channel.Response) (map[string]uint64, error) {
	if response.Payload == nil {
		return nil, nil
	}
	r := make(map[string]uint64)
	err := json.Unmarshal(response.Payload, &r)
	if err != nil {
		return nil, fmt.Errorf("unmarshal payload :%w", err)
	}

	return r, nil
}

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: plugins.Handshake,
		Plugins: map[string]plugin.Plugin{
			plugins.PluginName: &plugins.AppchainGRPCPlugin{Impl: &Client{}},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})

	logger.Info("Plugin server down")
}
