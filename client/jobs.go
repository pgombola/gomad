package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type details struct {
	Running int `json:"running"`
}

type summary struct {
	Details details `json:"server"`
}

type jobSummary struct {
	Summary summary `json:"Summary"`
}

type job struct {
	Name       string     `json:"Name"`
	Priority   int        `json:"Priority"`
	Status     string     `json:"status"`
	JobSummary jobSummary `json:"JobSummary"`
}

var jobs []job

func PopulateJobs(host string) {
	url := host + "/v1/jobs"

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	decodeJobs(resp.Body)
}

func PrintJobs() {
	for _, job := range jobs {
		fmt.Printf("Name=%v;Priority=%v;Status=%v;Running=%v\n", job.Name, job.Priority, job.Status, job.JobSummary.Summary.Details.Running)
	}
}

func decodeJobs(body io.ReadCloser) {
	decoder := json.NewDecoder(body)

	if err := decoder.Decode(&jobs); err != nil {
		log.Fatal(err)
	}
}
