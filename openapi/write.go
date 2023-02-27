package openapi

import (
	"RealtimeDB/rpc"
	"RealtimeDB/rtdb"
	"context"
	"encoding/json"
)
type Server struct {
	rpc.UnimplementedGreeterServer
}
func (s *Server) WritePoints(ctx context.Context, in *rpc.WritePointsRequest) (*rpc.WritePointsResponse, error){
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
	err = Store.InsertRows(rows)
	if err!=nil{
		return &rpc.WritePointsResponse{Reply: "error"}, nil
	}else{
		return &rpc.WritePointsResponse{Reply: "success"}, nil
	}
}
