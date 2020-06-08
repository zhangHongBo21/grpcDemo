/*
@Time : 2020/6/3 16:14
@Author : zhb
@File : client
@Software: GoLand
*/

package main

import (
	"context"
	"github.com/gogf/gf/os/glog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/grpclog"
	pb "grpcDemo/protoc"
	"log"
	"os"
)

type SearchService struct{}

func (s *SearchService) Search(ctx context.Context, r *pb.SearchRequest) (*pb.SearchResponse, error) {
	return &pb.SearchResponse{Query: r.GetQuery() + " Server"}, nil
}

// GetRequestMetadata 实现自定义认证接口
func (s *SearchService) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	return map[string]string{
		"appid":  "101010",
		"appkey": "i am key",
	}, nil
}
func (s *SearchService) RequireTransportSecurity() bool {
	return true
}

const port = "9001"

func tryCacthacc() {
	//捕获程序运行异常，防止程序意外退出
	if err := recover(); err != nil {
		glog.Notice(err)
		main()
	}
}
func main() {
	defer tryCacthacc()
	for {
		get()
	}
}

func get() {
	var opts []grpc.DialOption
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	creds, err := credentials.NewClientTLSFromFile(dir+"/keys/server.pem", "angang.zhs0909.com")
	if err != nil {
		grpclog.Fatalf("Failed to create TLS credentials %v", err)
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	opts = append(opts, grpc.WithPerRPCCredentials(new(SearchService)))
	conn, err := grpc.Dial(":"+port, opts...)

	if err != nil {
		log.Fatalf("grpc.Dial err: %v", err)
	}
	defer conn.Close()
	client := pb.NewSearchServiceClient(conn)
	resp, err := client.Search(context.Background(), &pb.SearchRequest{
		Query: "gRPC",
	})
	if err != nil {
		panic(err)
		//fmt.Sprintf("client.Search err: %v", err)
	}

	log.Printf("resp: %s", resp.GetQuery())
}
