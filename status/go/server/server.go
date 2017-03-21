package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/pgombola/gomad/client"
	pb "github.com/pgombola/gomad/status"
	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

var (
	tls = flag.Bool("tls", false, "True to use TLS for connection, false for plaintext.")
	// Add additional flags for certs
	port = flag.Int("port", 10000, "The server port.")
)

type statusError struct {
	S string
}

type clusterStatusServer struct {
}

func (s *clusterStatusServer) ListHosts(req *pb.HostsRequest, stream pb.ClusterStatus_ListHostsServer) error {
	nomad := &client.NomadServer{Address: "10.10.20.31", Port: 4646}
	hosts := client.Hosts(nomad)
	clarifyJob, _ := client.FindJob(nomad, "clarify")

	for _, host := range hosts {
		alloc, err := client.FindAlloc(nomad, clarifyJob, &host)
		if err != nil {
			// TODO clarify not allocated here
		}
		stream.Send(&pb.HostReply{Hostname: host.Name, Status: status(&host, clarifyJob, alloc)})
	}
	return nil
}

func status(host *client.Host, clarify *client.Job, alloc *client.Alloc) pb.HostReply_HostStatus {
	var status pb.HostReply_HostStatus
	fmt.Printf("Host: %v\n", host)
	fmt.Printf("Job: %v\n", clarify)
	fmt.Printf("Alloc: %v\n", alloc)
	if clarify.Name == "" {
		status = pb.HostReply_STOPPED
	} else if alloc.ClientStatus == "lost" && host.Drain {
		status = pb.HostReply_PENDING
	} else if alloc.ClientStatus == "running" && !alloc.CheckTaskStates("running") {
		status = pb.HostReply_MIXED
	} else {
		status = pb.HostReply_STARTED
	}
	fmt.Printf("Returning status: %v\n", status)
	return status
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
