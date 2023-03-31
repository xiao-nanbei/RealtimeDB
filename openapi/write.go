package openapi

import (
	"RealtimeDB/rpc"
	"context"
	"google.golang.org/grpc/peer"
)
type Server struct {
	rpc.UnimplementedGreeterServer
}
func (s *Server) WritePoints(ctx context.Context, in *rpc.WritePointsRequest) (*rpc.WritePointsResponse, error){
	p, _ := peer.FromContext(ctx)

	err := Store[Aps[p.Addr.String()]].WritePoints(in.Row)
	if err!=nil{
		return &rpc.WritePointsResponse{Reply: "error"}, nil
	}else{
		return &rpc.WritePointsResponse{Reply: "success"}, nil
	}
}
