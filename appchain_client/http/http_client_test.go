package main

import (
	"fmt"
	"testing"
)

func TestGetValue(t *testing.T) {
	httpClient := NewAppchainClient()
	fmt.Println((*httpClient).GetValue("5"))
}