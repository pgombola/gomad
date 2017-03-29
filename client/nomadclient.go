package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"io/ioutil"
	"bytes"
)

const http_bad_payload string = "400"
const http_unknown_error string = "520"

// NomadServer is connection parameters to a nomad server
type NomadServer struct {
	Address string
	Port    int
}

// Details represents a json object in a nomad job
type Details struct {
	Running int `json:"running"`
}

// Summary represents a json object in a nomad job
type Summary struct {
	Details *Details `json:"server"`
}

// JobSummary represents a json object in a nomad job
type JobSummary struct {
	Summary *Summary `json:"Summary"`
}

// Job is a representation of nomad job
type Job struct {
	Name       string      `json:"Name"`
	Priority   int         `json:"Priority"`
	Status     string      `json:"status"`
	JobSummary *JobSummary `json:"JobSummary"`
}

// Task represents an allocated nomad task
type Task struct {
	State string `json:"State"`
}

// Alloc is a representation of a nomad allocation
type Alloc struct {
	ID           string          `json:"ID"`
	JobID        string          `json:"JobID"`
	NodeID       string          `json:"NodeID"`
	Name         string          `json:"Name"`
	ClientStatus string          `json:"ClientStatus"`
	Tasks        map[string]Task `json:"TaskStates"`
}

// Host is a representation of a nomad client node
type Host struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Drain bool   `json:"Drain"`
}

// JobNotFound indicates that a Job search failed
type JobNotFound struct {
	Name string
}

// AllocNotFound indicates a missing allocation for a Job
type AllocNotFound struct {
	Jobname  string
	Hostname string
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

// Jobs will parse the json representation from the nomad rest api
// /v1/jobs
func Jobs(nomad *NomadServer) []Job {
	jobs := make([]Job, 0)
	decodeJSON(url(nomad)+"/v1/jobs", &jobs)
	return jobs
}

// FindJob will parse the json representation and find the supplied job name
func FindJob(nomad *NomadServer, name string) (*Job, error) {
	jobs := Jobs(nomad)
	for _, job := range jobs {
		if job.Name == name {
			return &job, nil
		}
	}
	return &Job{}, &JobNotFound{Name: name}
}

func (e *JobNotFound) Error() string {
	return fmt.Sprintf("Unable to find job name: %v", e.Name)
}

// Hosts will parse the json representation from the nomad rest api
// /v1/nodes
func Hosts(nomad *NomadServer) ([]Host, error) {
	hosts := make([]Host, 0)
	err := decodeJSON(url(nomad)+"/v1/nodes", &hosts)
	return hosts, err
}

// Drain will inform nomad to add/remove all allocations from that host
// depending on the value of enable
func Drain(nomad *NomadServer, id string, enable bool) (string, error) {
	resp, err := httpClient.Post(url(nomad)+"/v1/node/"+id+"/drain?enable="+strconv.FormatBool(enable), "application/json", nil)
	if resp != nil && resp.Body != nil {
	        defer resp.Body.Close()
	        return resp.Status, err
	}
	return http_unknown_error, err
}

func SubmitJob(nomad *NomadServer, launchFilePath string) (string, error) {
    file, err := ioutil.ReadFile(launchFilePath)
    if err != nil {
        return http_bad_payload, err
    }
    resp, err := httpClient.Post(url(nomad)+"/v1/jobs", "application/json", bytes.NewBuffer(file))
    if resp != nil && resp.Body != nil {
            defer resp.Body.Close()
            return resp.Status, err
    }
    return http_unknown_error, err
}

// Allocs will parse the json representation from the nomad rest api
// /v1/allocations
func Allocs(nomad *NomadServer) []Alloc {
	allocs := make([]Alloc, 0)
	decodeJSON(url(nomad)+"/v1/allocations", &allocs)
	return allocs
}

// FindAlloc will search through the Allocs on a provided Host to look for the
// allocations that match the provided Job
func FindAlloc(nomad *NomadServer, job *Job, host *Host) (*Alloc, error) {
	allocs := Allocs(nomad)
	for _, alloc := range allocs {
		if alloc.NodeID == host.ID && strings.Contains(alloc.Name, job.Name) {
			return &alloc, nil
		}
	}
	return &Alloc{}, &AllocNotFound{Hostname: host.Name, Jobname: job.Name}
}

// CheckTaskStates will return whether all of the tasks for an Alloc are
// in the state provided
func (alloc *Alloc) CheckTaskStates(state string) bool {
	for k := range alloc.Tasks {
		task := alloc.Tasks[k]
		if task.State != state {
			return false
		}
	}
	return true
}

func (e *AllocNotFound) Error() string {
	return fmt.Sprintf("Unable to find '%v' job on '%v' host.", e.Jobname, e.Hostname)
}

func url(nomad *NomadServer) string {
	return fmt.Sprintf("http://%v:%v", nomad.Address, nomad.Port)
}

func decodeJSON(url string, target interface{}) error {
	r, err := httpClient.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(target)
}
