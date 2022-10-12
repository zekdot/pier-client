package main

import (
	"github.com/jessevdk/go-flags"
)

type Init struct {
	Url     string `long:"url" description:"Specify URL of RPC API"`
}

func (args *Init) Name() string {
	return "init"
}

func (args *Init) UrlPassed() string {
	return args.Url
}

func (args *Init) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "init meta value", "Sends an dsswapper transaction to set <name> to <value>.", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *Init) Run() error {
	// Construct client
	url := args.Url
	if url == "" {
		url = RPC_URL
	}
	dsClient, err := GetClient(url)
	if err != nil {
		return err
	}
	err = dsClient.Init()
	return err
}
