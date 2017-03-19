package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	pb "github.com/pgombola/gomad/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	tls = flag.Bool("tls", false, "True to use TLS for connection, false for plaintext.")
	// Add additional flags for certs
	port = flag.Int("port", 10000, "The server port.")
)

type clusterStatusServer struct {
}

func (s *clusterStatusServer) ListHosts(req *pb.HostsRequest, stream pb.ClusterStatus_ListHostsServer) error {
	stream.Send(&pb.HostReply{Hostname: "server-1", Port: -1, Status: pb.HostReply_STARTED})
	return nil
}

func newServer() *clusterStatusServer {
	s := new(clusterStatusServer)
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		grpclog.Fatalf("Failed to listen: %v", err)
		os.Exit(-1)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterClusterStatusServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}
