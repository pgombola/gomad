package main

import (
	"fmt"

	"github.com/pgombola/gomad/client"
)

func main() {
	nodes := client.Status()
	for _, node := range nodes {
		fmt.Printf("ID=%v;Name=%v;Drain=%v\n", node.ID, node.Name, node.Drain)
	}
}
