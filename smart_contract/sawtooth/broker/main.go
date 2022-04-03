package main

import (
	"broker/contract"
	"broker/handler"
	"fmt"
	"github.com/hyperledger/sawtooth-sdk-go/logging"
	"github.com/hyperledger/sawtooth-sdk-go/processor"
	flags "github.com/jessevdk/go-flags"
	"os"
	"syscall"
)
type Opts struct {
	Verbose []bool `short:"v" long:"verbose" description:"Increase verbosity"`
	Connect string `short:"C" long:"connect" description:"Validator component endpoint to connect to" default:"tcp://localhost:4004"`
}

func main() {
	var opts Opts

	logger := logging.Get()

	parser := flags.NewParser(&opts, flags.Default)
	remaining, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		} else {
			logger.Errorf("Failed to parse args: %v", err)
			os.Exit(2)
		}
	}

	if len(remaining) > 0 {
		fmt.Printf("Error: Unrecognized arguments passed: %v\n", remaining)
		os.Exit(2)
	}

	endpoint := opts.Connect

	switch len(opts.Verbose) {
	case 2:
		logger.SetLevel(logging.DEBUG)
	case 1:
		logger.SetLevel(logging.INFO)
	default:
		logger.SetLevel(logging.WARN)
	}

	logger.Debugf("command line arguments: %v", os.Args)
	logger.Debugf("verbose = %v\n", len(opts.Verbose))
	logger.Debugf("endpoint = %v\n", endpoint)
	broker := contract.NewBroker()
	//fmt.Printf("create broker")
	handler := handler.NewHandler(broker)
	//fmt.Printf("create handler")
	processor := processor.NewTransactionProcessor(endpoint)
	processor.AddHandler(handler)
	//fmt.Printf("add handler")
	processor.ShutdownOnSignal(syscall.SIGINT, syscall.SIGTERM)
	//fmt.Printf("start processor")
	err = processor.Start()
	if err != nil {
		logger.Error("Processor stopped: ", err)
	}
}