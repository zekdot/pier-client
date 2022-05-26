package main


import (
	"fmt"
	"testing"
)

func TestFirstBundle(t *testing.T) {
	bm := NewBundleManager()
	bm.AddRequest("0x1223132465464", "mychannel&data_swapper", "key1")
	bm.AddRequest("0x1223132465464", "mychannel&data_swapper", "key2")
	bm.AddRequest("0x1223132465464", "mychannel&data_swapper", "key3")
	bm.AddRequest("0x1223132465464", "mychannel&data_swapper", "key4")
	bm.AddRequest("0x1223132465465", "mychannel&data_swapper", "key1")
	bm.AddRequest("0x1223132465464", "mychannel&data_swapper", "key5")
	bm.AddRequest("0x1223132465465", "mychannel&data_swapper", "key2")
	ev, err := bm.GetFirstBundle()
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println(ev)
	ev, err = bm.GetFirstBundle()
	fmt.Println(ev)
}
