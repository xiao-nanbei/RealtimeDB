package main

import (
	"RealtimeDB/openapi"
	"RealtimeDB/rpc"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)
var localhostPort string

func main() {
	file, err := os.ReadFile("start.txt")
	if err != nil {
		return
	}
	log.Println(string(file))
	go func() {
		var quit string
		for{
			fmt.Scanf("%s",&quit)
			if quit=="quit"{
				for _,v:=range openapi.Store{
					v.Close()
				}
				os.Exit(0)
			}
		}
	}()
	if err != nil {
		return
	}
	lis, err := net.Listen("tcp", ":8086")
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
		return
	}
	s := grpc.NewServer()                  // 创建gRPC服务器
	rpc.RegisterGreeterServer(s, &openapi.Server{}) // 在gRPC服务端注册服务
	err = s.Serve(lis)
	if err != nil {
		fmt.Printf("failed to serve: %v", err)
		return
	}
	s.GracefulStop()
}