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
	hosts := client.PopulateHosts("http://10.10.20.31:4646")
	clarifyJob := findClarifyJob(client.PopulateJobs("http://10.10.20.31:4646"))

	for _, host := range hosts {
		stream.Send(&pb.HostReply{Hostname: host.Name, Status: status(clarifyJob)})
	}
	return nil
}

func findClarifyJob(jobs []client.Job) *client.Job {
	for _, job := range jobs {
		if job.Name == "clarify" {
			return &job
		}
	}
	return &client.Job{Status: "stopped"}
}

func status(job *client.Job) pb.HostReply_HostStatus {
	switch job.Status {
	case "running":
		return pb.HostReply_STARTED
	case "stopped":
		return pb.HostReply_STOPPED
	}
	return pb.HostReply_MIXED
}

func clarifyStarted(job client.Job) bool {
	return job.Status == "running"
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
