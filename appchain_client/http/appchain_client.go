package main

type AppchainClient interface {
	//GetValue(Key string) ([]byte, error)
	//SetValue(Key string, Value string) error
	SetValue(string, string) error
	GetValue(string) (string, error)
}