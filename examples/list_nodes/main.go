// Example: list cluster nodes.
//
// Run with:
//
//	PMG_HOST=https://pmg.example.com:8006 \
//	PMG_TOKEN='PMGAPIToken=root@pam!auto=...' \
//	go run ./examples/list_nodes
package main

import (
	"context"
	"fmt"
	"os"

	pmg "github.com/client-api/pmg-go"
)

func main() {
	host := os.Getenv("PMG_HOST")
	if host == "" {
		host = "https://localhost:8006"
	}
	cfg := pmg.NewConfiguration()
	cfg.Servers = append(pmg.ServerConfigurations{}, pmg.ServerConfiguration{URL: host + "/api2/json"})
	cfg.DefaultHeader["Authorization"] = os.Getenv("PMG_TOKEN")

	client := pmg.NewAPIClient(cfg)
	resp, _, err := client.NodesAPI.NodesGetNodes(context.Background()).Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, "list nodes:", err)
		os.Exit(1)
	}
	nodes := resp.GetData()
	fmt.Printf("Found %d node(s):\n", len(nodes))
	for _, n := range nodes {
		// Other products expose a slimmer Node shape; print verbatim.
		fmt.Printf("  - %+v\n", n)
	}
}
