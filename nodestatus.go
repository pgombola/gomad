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
		client.PopulateHosts("http://" + nomad)
		client.PrintHosts()
	}

	if jobs {
		client.PopulateJobs("http://" + nomad)
		client.PrintJobs()
	}
}
