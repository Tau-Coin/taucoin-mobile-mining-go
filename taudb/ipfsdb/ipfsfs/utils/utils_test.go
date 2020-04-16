package utils

import (
	"fmt"
	"testing"
)

func TestByteToPath(t *testing.T){
	strData := "QmNess72JMwUVYZvigYEy2gDyk5xtFNWTVLND2FVaL9bkh"
	strByte := []byte(strData)
	path, err:= ByteToPath(strByte)
	if err != nil{
		fmt.Println(err)
		t.Fatalf("Fail")
	}
	t.Logf("Success")
	fmt.Println(path)
}
