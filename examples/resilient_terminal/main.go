// Example: resilient terminal session with auto-reconnect.
//
// Run with:
//
//	PMG_HOST=https://pmg.example.com:8006 \
//	PMG_TOKEN='PMGAPIToken=root@pam!auto=...' \
//	PMG_NODE=orca PMG_VMID=100 \
//	go run ./examples/resilient_terminal
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
	opts := pmg.RetryOptions{
		MaxRetries:        20,
		InitialDelay:      250 * time.Millisecond,
		MaxDelay:          30 * time.Second,
		BackoffMultiplier: 2.0,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	session, err := client.ConnectTerminalResilient(ctx, target, opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect:", err)
		os.Exit(1)
	}
	session.OnMessage = func(text string) { fmt.Print(text) }
	session.OnReconnect = func(attempt int) { fmt.Printf("\n[reconnected after %d attempts]\n", attempt) }
	session.OnGiveUp = func(err error) { fmt.Printf("\n[retries exhausted: %v]\n", err) }

	_ = session.Send("date\n")
	deadline := time.Now().Add(5 * time.Minute)
	tick := time.NewTicker(30 * time.Second)
	defer tick.Stop()
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			break
		case <-tick.C:
			_ = session.Send("date\n")
		}
	}
	_ = session.Close()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
