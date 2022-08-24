package main

import (
	"fmt"
	"testing"
)

func TestGetValue(t *testing.T) {
	httpClient := NewHttpClient()
	fmt.Println(httpClient.GetValue("5"))
}