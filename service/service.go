/*
@Time : 2020/6/3 16:14
@Author : zhb
@File : service
@Software: GoLand
*/
package main

import (
	"context"
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/metadata"
	pb "grpcDemo/protoc"
	"log"
	"net"
	"os"
)

type SearchService struct{}

func (s *SearchService) Search(ctx context.Context, r *pb.SearchRequest) (*pb.SearchResponse, error) {
	e, b := sentinel.Entry("some-test")
	if b != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "超过限流")
	}
	defer e.Exit()
	// 解析metada中的信息并验证
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
	}
	var (
		appid  string
		appkey string
	)
	if val, ok := md["appid"]; ok {
		appid = val[0]
	}
	if val, ok := md["appkey"]; ok {
		appkey = val[0]
	}
	if appid != "101010" || appkey != "i am key" {
		return nil, grpc.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
	}

	return &pb.SearchResponse{Query: r.GetQuery() + " Server"}, nil
}

const port = "9001"

func main() {
	// 务必先进行初始化
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatal(err)
	}
	// 配置一条限流规则
	_, err = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "some-test",
			MetricType:      flow.QPS,
			Count:           10,
			ControlBehavior: flow.Reject,
		},
	})
	if err != nil {
		fmt.Println(err)
		return
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	// TLS认证
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	creds, err := credentials.NewServerTLSFromFile(dir+"/keys/server.pem", dir+"/keys/server.key")
	if err != nil {
		grpclog.Fatalf("Failed to generate credentials %v", err)
	}
	// 实例化grpc Server, 并开启TLS认证
	grpcServer := grpc.NewServer(grpc.Creds(creds))

	//grpcServer := grpc.NewServer()

	pb.RegisterSearchServiceServer(grpcServer, &SearchService{})
	grpcServer.Serve(lis)
}
