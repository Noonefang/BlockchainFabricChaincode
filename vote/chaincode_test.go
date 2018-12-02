package main

import (
	"testing"
	"fmt"
	"bytes"
)

func TestByte(t *testing.T) {
	a:=[]byte{0,1,2,3,4}
	b:=[]byte{0,1,2,3,8}
	fmt.Println(bytes.Equal(a,b))
}
