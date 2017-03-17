package client

import (
	"encoding/json"
	"fmt"
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

var nodes []Node

func PopulateHosts(host string) {
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
	decodeNodes(resp.Body)
}

func PrintHosts() {
	for _, node := range nodes {
		fmt.Printf("ID=%v;Name=%v;Drain=%v\n", node.ID, node.Name, node.Drain)
	}
}

func decodeNodes(body io.ReadCloser) {
	decoder := json.NewDecoder(body)

	if err := decoder.Decode(&nodes); err != nil {
		log.Fatal(err)
	}
}
