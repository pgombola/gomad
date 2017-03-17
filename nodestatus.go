package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pgombola/gomad/client"
)

var (
	nomad string
	hosts bool
	jobs  bool
)

func init() {
	flag.StringVar(&nomad, "nomad", "", "Host address and port of nomad server")
	flag.BoolVar(&hosts, "hosts", false, "Retrieve the status of the hosts")
	flag.BoolVar(&jobs, "jobs", false, "Retrieve the status of the jobs")
}

func main() {
	flag.Parse()

	if nomad == "" {
		fmt.Println("nomad flag must be set.")
		os.Exit(-1)
	}

	if hosts {
		nodes := client.Status("http://" + nomad)
		for _, node := range nodes {
			fmt.Printf("ID=%v;Name=%v;Drain=%v\n", node.ID, node.Name, node.Drain)
		}
	}

	if jobs {
		jobs := client.Jobs("http://" + nomad)
		for _, job := range jobs {
			fmt.Printf("Name=%v;Priority=%v;Status=%v;Running=%v\n", job.Name, job.Priority, job.Status, job.JobSummary.Summary.Details.Running)
		}
	}
}
