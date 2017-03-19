package main

import (
	"context"
	"flag"
	"io"
	"os"

	pb "github.com/pgombola/gomad/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	tls        = flag.Bool("tls", false, "True to use TLS for connection, false for plaintext.")
	serverAddr = flag.String("server_addr", "127.0.0.1:10000", "The server address in host:port format.")
)

func printHosts(client pb.ClusterStatusClient) {
	grpclog.Printf("Retrieving hosts...")
	stream, err := client.ListHosts(context.Background(), &pb.HostsRequest{})

	if err != nil {
		grpclog.Fatalf("%v.ListHosts(_) = _, %v", client, err)
	}
	for {
		host, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			grpclog.Fatalf("%v.ListHosts(_) = _, %v", client, err)
		}
		grpclog.Println(host)
	}
}

func main() {
	flag.Parse()
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	conn, err := grpc.Dial(*serverAddr, opts...)
	if err != nil {
		grpclog.Fatalf("fail to dial: %v", err)
		os.Exit(-1)
	}
	defer conn.Close()
	client := pb.NewClusterStatusClient(conn)

	printHosts(client)
}
