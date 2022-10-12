package main

import (
	"github.com/jessevdk/go-flags"
)

type Set struct {
	Args struct {
		Name  string `positional-arg-name:"name" required:"true" description:"Name of key to set"`
		Value string `positional-arg-name:"value" required:"true" description:"Amount to set"`
	} `positional-args:"true"`
	Url string `long:"url" description:"Specify URL of RPC API"`
}

func (args *Set) Name() string {
	return "set"
}

func (args *Set) UrlPassed() string {
	return args.Url
}

func (args *Set) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "Sets an dsswapper value", "Sends an dsswapper transaction to set <name> to <value>.", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *Set) Run() error {
	// Construct client
	name := args.Args.Name
	value := args.Args.Value
	url := args.Url
	if url == "" {
		url = RPC_URL
	}
	dsClient, err := GetClient(url)
	if err != nil {
		return err
	}
	err = dsClient.SetData(name, value)
	return err
}
