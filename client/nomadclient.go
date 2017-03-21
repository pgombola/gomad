package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
	"strconv"
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
	Details Details `json:"server"`
}

// JobSummary represents a json object in a nomad job
type JobSummary struct {
	Summary Summary `json:"Summary"`
}

// Job is a representation of nomad job
type Job struct {
	Name       string     `json:"Name"`
	Priority   int        `json:"Priority"`
	Status     string     `json:"status"`
	JobSummary JobSummary `json:"JobSummary"`
}

// Host is a representation of a nomad client node
type Host struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Drain bool   `json:"Drain"`
}

var httpClient = &http.Client{Timeout: 5 * time.Second}

// Jobs will parse the json representation from the nomad rest api
func Jobs(nomad *NomadServer) []Job {
	jobs := make([]Job, 0)
	decodeJSON(url(nomad)+"/v1/jobs", &jobs)
	return jobs
}

// Hosts will parse the json representation from the nomad rest api
func Hosts(nomad *NomadServer) []Host {
	hosts := make([]Host, 0)
	decodeJSON(url(nomad)+"/v1/nodes", &hosts)
	return hosts
}

func Drain(nomad *NomadServer, id string, enable bool) string {
        client := &http.Client{}
        resp, _ := client.Post(url(nomad) + "/v1/node/" + id + "/drain?enable=" + strconv.FormatBool(enable), "application/json", nil)
        return resp.Status
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
