package client

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

type Details struct {
	Running int `json:"running"`
}

type Summary struct {
	Details Details `json:"server"`
}

type JobSummary struct {
	Summary Summary `json:"Summary"`
}

type Job struct {
	Name       string     `json:"Name"`
	Priority   int        `json:"Priority"`
	Status     string     `json:"status"`
	JobSummary JobSummary `json:"JobSummary"`
}

func PopulateJobs(host string) []Job {
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
	return decodeJobs(resp.Body)
}

func PrintJobs(jobs []Job) {
	for _, job := range jobs {
		fmt.Printf("Name=%v;Priority=%v;Status=%v;Running=%v\n", job.Name, job.Priority, job.Status, job.JobSummary.Summary.Details.Running)
	}
}

func decodeJobs(body io.ReadCloser) []Job {
	decoder := json.NewDecoder(body)

	var jobs []Job
	if err := decoder.Decode(&jobs); err != nil {
		log.Fatal(err)
	}
	return jobs
}
