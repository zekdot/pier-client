package main

import (
	"fmt"
	"github.com/jessevdk/go-flags"
)

type Get struct {
	Args struct {
		Name string `positional-arg-name:"name" required:"true" description:"Name of key to show"`
	} `positional-args:"true"`
	Url string `long:"url" description:"Specify URL of REST API"`
}

func (args *Get) Name() string {
	return "get"
}

func (args *Get) KeyfilePassed() string {
	return ""
}

func (args *Get) UrlPassed() string {
	return args.Url
}

func (args *Get) Register(parent *flags.Command) error {
	_, err := parent.AddCommand(args.Name(), "Displays the specified intkey value", "Shows the value of the key <name>.", args)
	if err != nil {
		return err
	}
	return nil
}

func (args *Get) Run() error {
	// Construct client
	name := args.Args.Name
	dsClient, err := GetClient(args, false)
	if err != nil {
		return err
	}
	value, err := dsClient.GetData(name)
	if err != nil {
		return err
	}
	fmt.Println(name, ": ", value)
	return nil
}