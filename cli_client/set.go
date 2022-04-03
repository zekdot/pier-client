package main

import (
	"github.com/jessevdk/go-flags"
)

type Set struct {
	Args struct {
		Name  string `positional-arg-name:"name" required:"true" description:"Name of key to set"`
		Value string `positional-arg-name:"value" required:"true" description:"Amount to set"`
	} `positional-args:"true"`
	Url     string `long:"url" description:"Specify URL of REST API"`
	Keyfile string `long:"keyfile" description:"Identify file containing user's private key"`
	Wait    uint   `long:"wait" description:"Set time, in seconds, to wait for transaction to commit"`
}

func (args *Set) Name() string {
	return "set"
}

func (args *Set) KeyfilePassed() string {
	return args.Keyfile
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
	//wait := args.Wait

	dsClient, err := GetClient(args, true)
	if err != nil {
		return err
	}
	err = dsClient.SetData(name, value)
	return err
}
