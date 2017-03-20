package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
	"strconv"
)

type Host struct {
	ID    string `json:"ID"`
	Name  string `json:"Name"`
	Drain bool   `json:"Drain"`
}

func PopulateHosts(host string) []Host {
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
	return decodeNodes(resp.Body)
}

func Drain(host string, id string, enabled bool) {
    url := host + "/v1/node/" + id + "/drain?enable=" + strconv.FormatBool(enabled)
        
    client := &http.Client{}
    resp, _ := client.Post(url, "application/json", nil)
    fmt.Println(resp.Status)
}

func HostsToString(hosts []Host) string {
	var buffer bytes.Buffer
	for _, host := range hosts {
		fmt.Fprintf(&buffer, "ID=%v;Name=%v;Drain=%v\n", host.ID, host.Name, host.Drain)
	}
	return buffer.String()
}

func decodeNodes(body io.ReadCloser) []Host {
	decoder := json.NewDecoder(body)
	var hosts []Host
	if err := decoder.Decode(&hosts); err != nil {
		log.Fatal(err)
	}
	return hosts
}
