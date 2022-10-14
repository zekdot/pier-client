package main


import (
	"fmt"
	"strconv"
	"testing"
)

func TestFirstBundle(t *testing.T) {
	bm := NewBundleManager()
	for i := 1; i <= 100; i ++ {
		bm.AddRequest("0x1223132465464", "mychannel&data_swapper", "key" + strconv.FormatInt(int64(i), 10))
	}
	ev, err := bm.GetFirstBundle()
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println(ev)

	for i := 1; i <= 100; i ++ {
		bm.AddRequest("0x1223132465464", "mychannel&data_swapper", "key" + strconv.FormatInt(int64(i), 10))
	}
	ev, err = bm.GetFirstBundle()
	if err != nil {
		t.Error(err.Error())
	}
	fmt.Println(ev)
	ev, err = bm.GetFirstBundle()
	fmt.Println(ev)
}
