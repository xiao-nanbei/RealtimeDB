package tests

import (
	"RealtimeDB/openapi"
	"testing"
)
func Test_WRITE(t *testing.T) {
	openapi.Config("./testdata")
	dataAll:=make([][]float64,0)
	for i:=0;i<1000;i++{
		datas:=make([]float64,8)
		for j:=0;j<len(datas);j++{
			datas[j]=float64(i)
		}
		dataAll=append(dataAll,datas)
	}
	for i:=0;i<1000;i++{
		err := openapi.Write(dataAll[i], "host1", "core1", "process1")
		if err != nil {
			return
		}
	}

}

