package main

import (
	"github.com/jessevdk/go-flags"
)

type Init struct {
	//Args struct {
	//	Name  string `positional-arg-name:"name" required:"true" description:"Name of key to set"`
	//	Value string `positional-arg-name:"value" required:"true" description:"Amount to set"`
	//} `positional-args:"true"`
	Url     string `long:"url" description:"Specify URL of REST API"`
	Keyfile string `long:"keyfile" description:"Identify file containing user's private key"`
	Wait    uint   `long:"wait" description:"Set time, in seconds, to wait for transaction to commit"`
}

func (args *Init) Name() string {
	return "init"
}

func (args *Init) KeyfilePassed() string {
	return args.Keyfile
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
	//name := args.Args.Name
	//value := args.Args.Value
	//wait := args.Wait

	dsClient, err := GetClient()
	if err != nil {
		return err
	}
	err = dsClient.Init()
	return err
}
