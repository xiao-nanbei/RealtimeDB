package main

import (
	"RealtimeDB/openapi"
	"RealtimeDB/rpc"
	"RealtimeDB/rtdb"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"strings"
)

type server struct {
	rpc.UnimplementedGreeterServer
}
func (s *server) WritePoints(ctx context.Context, in *rpc.WritePointsRequest) (*rpc.WritePointsResponse, error){
	metirc_tags:=strings.Split(in.MetricTags," ")
	res:=make([]string,0)
	for _,metirc_tag:=range metirc_tags{
		res=append(res,strings.Split(metirc_tag,":")...)
	}
	err:=openapi.Write([]float64{in.Data},res)
	if err!=nil{
		return &rpc.WritePointsResponse{Reply: "error"}, nil
	}else{
		return &rpc.WritePointsResponse{Reply: "success"}, nil
	}
}
func main() {
	file, err := os.ReadFile("start.txt")
	if err != nil {
		return
	}
	fmt.Println(string(file))
	store := rtdb.OpenRTDB()
	defer store.Close()
	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()                  // 创建gRPC服务器
	rpc.RegisterGreeterServer(s, &server{}) // 在gRPC服务端注册服务
	err = s.Serve(lis)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
}