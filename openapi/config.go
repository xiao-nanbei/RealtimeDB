package openapi

import (
	"RealtimeDB/rtdb"
	"github.com/chenjiandongx/logger"
)


var Store rtdb.RTDB
func Config(path string){
	Store = *rtdb.OpenRTDB(rtdb.WithDataPath(path), rtdb.WithLoggerConfig(&logger.Options{
		Stdout:      true,
		ConsoleMode: true,
		Level:       logger.ErrorLevel,
	}))

}
