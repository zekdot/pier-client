package main

import (
	"github.com/jessevdk/go-flags"
)

type InterchainGet struct {
	Args struct {
		TargetChainId  string `positional-arg-name:"tid" required:"true" description:"target pier id"`
		CCid string `positional-arg-name:"ccid" required:"true" description:"target contract's address"`
		Key string `positional-arg-name:"key" required:"true" description:"value's key"`
	} `positional-args:"true"`
	Url     string `long:"url" description:"Specify URL of RPC API"`
}

func (args *InterchainGet) Name() string {
	return "interchainGet"
}

func (args *InterchainGet) UrlPassed() string {
	return args.Url
}

func (args *InterchainGet) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "ff", "ff", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *InterchainGet) Run() error {
	// Construct client
	toId := args.Args.TargetChainId
	cid := args.Args.CCid
	key := args.Args.Key

	dsClient, err := GetClient(args.Url)
	if err != nil {
		return err
	}
	err = dsClient.InterchainGet(toId, cid, key)
	return err
}