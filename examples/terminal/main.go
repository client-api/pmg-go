// Example: open a terminal session against a QEMU VM.
//
// Run with:
//
//	PMG_HOST=https://pmg.example.com:8006 \
//	PMG_TOKEN='PMGAPIToken=root@pam!auto=...' \
//	PMG_NODE=orca PMG_VMID=100 \
//	go run ./examples/terminal
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	pmg "github.com//"
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
	node := envOr("PMG_NODE", "pmg1")
	vmid64, _ := strconv.ParseInt(envOr("PMG_VMID", "100"), 10, 32)
	vmid := int32(vmid64)

	target := pmg.Target{Kind: pmg.TargetKindQemu, Node: node, Vmid: vmid}
	fmt.Printf("Opening terminal on %s:qemu/%d...\n", node, vmid)

	session, err := client.ConnectTerminal(context.Background(), target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	session.OnMessage = func(text string) { fmt.Print(text) }
	session.OnClose = func(err error) { fmt.Printf("\n[closed: %v]\n", err) }

	if err := session.Resize(120, 32); err != nil {
		fmt.Fprintln(os.Stderr, "resize:", err)
	}
	if err := session.Send("uname -a\n"); err != nil {
		fmt.Fprintln(os.Stderr, "send:", err)
	}
	time.Sleep(5 * time.Second)
	_ = session.Close()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
