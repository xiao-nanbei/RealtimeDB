package openapi

import (
	"RealtimeDB/rpc"
	"RealtimeDB/rtdb"
	"context"
	"github.com/chenjiandongx/logger"
)
var Store rtdb.RTDB
func (s *Server)Config(ctx context.Context, in *rpc.ConfigRequest) (*rpc.ConfigResponse, error){
	Store = *rtdb.OpenRTDB(rtdb.WithDataPath(in.Path), rtdb.WithLoggerConfig(&logger.Options{
		Stdout:      true,
		ConsoleMode: true,
		Level:       logger.ErrorLevel,
	}))
	return &rpc.ConfigResponse{Reply: "success"}, nil
}
