package main

import (
	"RealtimeDB/client/rpc"
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"
)
var addressPort string
func main() {
	flag.Parse()
	file, err := os.ReadFile("start.txt")
	if err != nil {
		return
	}
	log.Println(string(file))
	log.Println("please enter the address and port:")
	_, err = fmt.Scanf("%s", &addressPort)
	if err != nil {
		return
	}
	addr := flag.String("addr",addressPort , "the address to connect to")
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := rpc.NewGreeterClient(conn)
	// 执行RPC调用并打印收到的响应数据

	log.Println("please enter the data path:")
	var path string
	_, err = fmt.Scanf("%s", &path)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()
	r,err := c.Config(ctx,&rpc.ConfigRequest{Path: path})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetReply())
	exitflag:=false
	for {
		var enter string
		fmt.Scanf("%s",&enter)
		switch enter {
			case "write":
				var row string
				fmt.Scanf("%s",&row)
				r, err := c.WritePoints(ctx, &rpc.WritePointsRequest{Row: row})
				if err != nil {
					log.Fatalf("could not greet: %v", err)
				}
				log.Printf("Greeting: %s", r.GetReply())
				break
			case "queryseries":
				var tags string
				fmt.Scanf("%s",&tags)
				r, err := c.QuerySeries(ctx, &rpc.QuerySeriesRequest{Tags: tags})
				if err != nil {
					log.Fatalf("could not greet: %v", err)
				}
				log.Printf("Greeting: %s", r.GetReply())
				break
			case "queryrange":
				var metric_tags string
				fmt.Scanf("%s",&metric_tags)
				r, err := c.QueryRange(ctx, &rpc.QueryRangeRequest{MetricTags: metric_tags})
				if err != nil {
					log.Fatalf("could not greet: %v", err)
				}
				log.Printf("Greeting: %s", r.GetReply())
				break
			case "querytagvalues":
				var tag string
				fmt.Scanf("%s",&tag)
				r, err := c.QueryTagValues(ctx, &rpc.QueryTagValuesRequest{Tag: tag})
				if err != nil {
					log.Fatalf("could not greet: %v", err)
				}
				log.Printf("Greeting: %s", r.GetReply())
				break
			case "exit":
				exitflag=true
				break
			default:
				log.Println("error enter")
				break
		}
		if exitflag==true{
			break
		}
	}
}
