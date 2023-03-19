package openapi

import (
	"RealtimeDB/rpc"
	"RealtimeDB/rtdb"
	"context"
	"github.com/chenjiandongx/logger"
	"google.golang.org/grpc/peer"
)
var Aps map[string]string = make(map[string]string,0)
var Store map[string]*rtdb.RTDB = make(map[string]*rtdb.RTDB,0)
var TestStore rtdb.RTDB
func (s *Server)Config(ctx context.Context, in *rpc.ConfigRequest) (*rpc.ConfigResponse, error){
	p, _ := peer.FromContext(ctx)
	Aps[p.Addr.String()]=in.Name
	if _,ok:=Store[in.Name];ok{
		return &rpc.ConfigResponse{Reply: "success"}, nil
	}else{
		Store[in.Name] = rtdb.OpenRTDB(rtdb.WithDataPath("./data/"+in.Name), rtdb.WithLoggerConfig(&logger.Options{
			Stdout:      true,
			ConsoleMode: true,
			Level:       logger.ErrorLevel,
		}))

	}
	return &rpc.ConfigResponse{Reply: "success"}, nil
}
