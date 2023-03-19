package openapi

import (
	"RealtimeDB/rpc"
	"RealtimeDB/rtdb"
	"context"
	"encoding/json"
	"github.com/chenjiandongx/logger"
	"google.golang.org/grpc/peer"
)
type Server struct {
	rpc.UnimplementedGreeterServer
}
func (s *Server) WritePoints(ctx context.Context, in *rpc.WritePointsRequest) (*rpc.WritePointsResponse, error){
	p, _ := peer.FromContext(ctx)
	var slice map[string]interface{}
	err := json.Unmarshal([]byte(in.Row), &slice)
	if err != nil {
		return nil, err
	}
	row:=&rtdb.Row{
		Metric: slice["metric"].(string),
		Point: rtdb.Point{
			TimeStamp: int64(slice["timestamp"].(float64)),
			Value: slice["value"].(float64),
		},
	}
	delete(slice,"metric")
	delete(slice,"timestamp")
	delete(slice,"value")
	for k,v:=range slice{
		row.Tags=append(row.Tags,rtdb.Tag{Name: k,Value: v.(string)})
	}
	rows := []*rtdb.Row{row}
	if _,ok:=Store[Aps[p.Addr.String()]];!ok{
		Aps[p.Addr.String()]="testing"
		Store["testing"] = rtdb.OpenRTDB(rtdb.WithDataPath("./data/testing"), rtdb.WithLoggerConfig(&logger.Options{
			Stdout:      true,
			ConsoleMode: true,
			Level:       logger.ErrorLevel,
		}))
	}
	err = Store[Aps[p.Addr.String()]].InsertRows(rows)
	if err!=nil{
		return &rpc.WritePointsResponse{Reply: "error"}, nil
	}else{
		return &rpc.WritePointsResponse{Reply: "success"}, nil
	}
}
func WritePoints(r string)error{
	var slice map[string]interface{}
	err := json.Unmarshal([]byte(r), &slice)
	if err != nil {
		return err
	}
	row:=&rtdb.Row{
		Metric: slice["metric"].(string),
		Point: rtdb.Point{
			TimeStamp: int64(slice["timestamp"].(float64)),
			Value: slice["value"].(float64),
		},
	}
	delete(slice,"metric")
	delete(slice,"timestamp")
	delete(slice,"value")
	for k,v:=range slice{
		row.Tags=append(row.Tags,rtdb.Tag{Name: k,Value: v.(string)})
	}
	rows := []*rtdb.Row{row}
	err = TestStore.InsertRows(rows)
	return nil
}