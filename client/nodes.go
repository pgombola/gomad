package client

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

type Node struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Drain bool   `json:"Drain"`
}

func Status(host string) []Node {
	url := host + "/v1/nodes"

	timeout := time.Duration(5 * time.Second)
	client := http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	return decode(resp.Body)
}

func decode(body io.ReadCloser) []Node {
	decoder := json.NewDecoder(body)

	var nodes []Node
	if err := decoder.Decode(&nodes); err != nil {
		log.Fatal(err)
	}
	return nodes
}
