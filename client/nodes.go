package client

import (
	"encoding/json"
	"log"
	"net/http"
)

type Node struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Drain bool   `json:"Drain"`
}

func Status() []Node {
	url := "http://10.10.20.32:4646/v1/nodes"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)

	var nodes []Node
	if err := decoder.Decode(&nodes); err != nil {
		log.Fatal(err)
	}
	return nodes
}
