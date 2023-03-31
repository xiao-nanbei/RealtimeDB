package tests

import (
	"RealtimeDB/openapi"
	"RealtimeDB/rtdb"
	"bufio"
	"fmt"
	"github.com/chenjiandongx/logger"
	"log"
	"os"
	"testing"
	"time"
)

func Test_Write(t *testing.T){
	openapi.TestStore = *rtdb.OpenRTDB(rtdb.WithDataPath("./testdata"), rtdb.WithLoggerConfig(&logger.Options{
		Stdout:      true,
		ConsoleMode: true,
		Level:       logger.ErrorLevel,
	}))
	f,err:=os.Open("/home/databrains/data.txt")
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer f.Close()
	log.Println(time.Now())
	br := bufio.NewReader(f)
	for {
		a, _, _ := br.ReadLine()
		openapi.TestStore.WritePoints(string(a))
	}
}
