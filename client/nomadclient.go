package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	kitlog "github.com/go-kit/kit/log"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

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
	JobID   string             `json:"JobID"`
	Summary map[string]Details `json:"Summary"`
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
	ID            string          `json:"ID"`
	JobID         string          `json:"JobID"`
	NodeID        string          `json:"NodeID"`
	Name          string          `json:"Name"`
	ClientStatus  string          `json:"ClientStatus"`
	DesiredStatus string          `json:"DesiredStatus"`
	TaskGroup     string          `json:"TaskGroup"`
	Tasks         map[string]Task `json:"TaskStates"`
}

// Host is a representation of a nomad client node
type Host struct {
	ID        string         `json:"ID"`
	Name      string         `json:"Name"`
	Drain     bool           `json:"Drain"`
	Resources *NodeResources `json:"Resources,omitempty"`
}

// NodeResources represents the resources of a nomad node
type NodeResources struct {
	Networks *Networks `json:"Networks"`
}

// Networks represents the network available on a nomad node
type Networks struct {
	IP string `json:"IP"`
}

// AllocNotFound indicates a missing allocation for a Job
type AllocNotFound struct {
	Jobname  string
	Hostname string
}

func log(keyvals ...interface{}) *bytes.Buffer {
	w := new(bytes.Buffer)
	kitlog.NewJSONLogger(w).Log(keyvals)
	return w
}

// Jobs will parse the json representation from the nomad rest api
// /v1/jobs
func Jobs(nomad *NomadServer) ([]Job, int, error) {
	jobs := make([]Job, 0)
	status, err := decodeJSON(url(nomad)+"/v1/jobs", &jobs)
	return jobs, status, err
}

// FindJob will parse the json representation and find the supplied job name
func FindJob(nomad *NomadServer, name string) (*Job, error) {
	jobs, _, _ := Jobs(nomad)
	for _, job := range jobs {
		if job.Name == name {
			return &job, nil
		}
	}
	return &Job{}, errors.New("job not found")
}

// Hosts will parse the json representation from the nomad rest api
// /v1/nodes
func Hosts(nomad *NomadServer) ([]Host, int, error) {
	hosts := make([]Host, 0)
	status, err := decodeJSON(url(nomad)+"/v1/nodes", &hosts)
	return hosts, status, err
}

// Drain will inform nomad to add/remove all allocations from that host
// depending on the value of enable
// Returns the http status code or an error
func Drain(nomad *NomadServer, id string, enable bool) (int, error) {
	resp, err := retryablehttp.Post(url(nomad)+"/v1/node/"+id+"/drain?enable="+strconv.FormatBool(enable), "application/json", nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		return resp.StatusCode, err
	}
	return http.StatusInternalServerError, err
}

// StopJob will send a delete request to a nomad server to stop the provided job
func StopJob(nomad *NomadServer, job *Job) (int, error) {
	url := fmt.Sprintf("http://%v:%v/v1/job/%v", nomad.Address, nomad.Port, job.Name)
	req, err := retryablehttp.NewRequest(http.MethodDelete, url, bytes.NewReader([]byte{}))
	if err != nil {
		buf := log("event", "stop_job_request_error", "jobname", job.Name, "error", err.Error())
		return http.StatusInternalServerError, errors.New(buf.String())
	}
	resp, err := retryablehttp.NewClient().Do(req)
	if err != nil {
		buf := log("event", "stop_job_client_error", "jobname", job.Name, "error", err.Error())
		return resp.StatusCode, errors.New(buf.String())
	}
	return resp.StatusCode, nil
}

// SubmitJob requests that nomade run a json launch file located at the supplied path
// Returns an http status code or error
func SubmitJob(nomad *NomadServer, launchFilePath string) (int, error) {
	file, err := ioutil.ReadFile(launchFilePath)
	if err != nil {
		return http.StatusBadRequest, err
	}
	resp, err := retryablehttp.Post(url(nomad)+"/v1/jobs", "application/json", bytes.NewReader(file))
	defer resp.Body.Close()
	return resp.StatusCode, err
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
			// We may be looking at a stale allocation and a newer one exists
			if alloc.DesiredStatus == "stop" && len(allocs) > 1 {
				continue
			}
			return &alloc, nil
		}
	}
	return &Alloc{}, &AllocNotFound{Hostname: host.Name, Jobname: job.Name}
}

//Active determines whether a given host is active.  This is true when we find a worker
//on a given host that isn't in a "stop" desired status
func Active(nomad *NomadServer, job *Job, host *Host) bool {
	allocs := Allocs(nomad)
	for _, alloc := range allocs {
		if alloc.NodeID == host.ID && strings.Contains(alloc.Name, job.Name) {
			if alloc.DesiredStatus != "stop" && strings.Contains(alloc.Name, "worker") {
				return true
			}
		}
	}
	return false
}

//Services returns a list of services present on a given host
func Services(nomad *NomadServer, job *Job, host *Host) []string {
	services := make([]string, 0)
	allocs := Allocs(nomad)
	for _, alloc := range allocs {
		if alloc.NodeID == host.ID && strings.Contains(alloc.Name, job.Name) {
			if alloc.DesiredStatus != "stop" {
				services = append(services, alloc.TaskGroup)
			}
		}
	}
	return services
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

// HostID finds a nomad Host that matches the provided hostname
func HostID(nomad *NomadServer, hostname *string) (*Host, error) {
	hosts, _, err := Hosts(nomad)
	if err != nil {
		return &Host{}, err
	}
	for _, host := range hosts {
		if *hostname == host.Name {
			return &host, nil
		}
	}
	buf := log("event", "node_not_found", "hostname", hostname)
	return &Host{}, errors.New(buf.String())

}

func (e *AllocNotFound) Error() string {
	return fmt.Sprintf("Unable to find '%v' job on '%v' host.", e.Jobname, e.Hostname)
}

func url(nomad *NomadServer) string {
	return fmt.Sprintf("http://%v:%v", nomad.Address, nomad.Port)
}

func decodeJSON(url string, target interface{}) (int, error) {
	r, err := retryablehttp.Get(url)
	if err != nil && r != nil && r.Body != nil {
		defer r.Body.Close()
		return r.StatusCode, err
	} else if err != nil {
		return http.StatusInternalServerError, err
	}
	return http.StatusOK, json.NewDecoder(r.Body).Decode(target)
}
