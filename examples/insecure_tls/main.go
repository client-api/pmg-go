// Example: connect to a Proxmox host with a self-signed certificate.
//
// The PVE web UI ships with a self-signed cert by default. Production
// setups should use a real CA-signed cert (Let's Encrypt via the
// Proxmox UI), but home-lab and dev setups commonly need to opt out
// of cert verification.
//
// **Security note:** disabling verification is vulnerable to MITM.
// Use only on trusted networks, or pin a custom CertPool instead.
//
// Run with:
//
//	PMG_HOST=https://pmg.example.com:8006 \
//	PMG_TOKEN='PMGAPIToken=root@pam!auto=...' \
//	PMG_NODE=orca PMG_VMID=100 \
//	go run ./examples/insecure_tls
package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	pmg "github.com//"
	gws "github.com/gorilla/websocket"
)

func main() {
	host := envOr("PMG_HOST", "https://localhost:8006")

	// ── 1. REST: install a Transport that skips cert verification.
	insecureTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
	}

	cfg := pmg.NewConfiguration()
	cfg.Servers = pmg.ServerConfigurations{
		pmg.ServerConfiguration{URL: host + "/api2/json"},
	}
	cfg.DefaultHeader["Authorization"] = os.Getenv("PMG_TOKEN")
	cfg.HTTPClient = &http.Client{Transport: insecureTransport}

	// ── 2. WebSocket: replace the package-level DefaultTransport with
	//    a GorillaTransport whose Dialer carries the same insecure
	//    TLSClientConfig. This is resolved at connect-time, so it
	//    must be set before the first ConnectTerminal / ConnectVnc
	//    call (here, before NewAPIClient is even fine).
	pmg.DefaultTransport = pmg.GorillaTransport{
		Dialer: &gws.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}, //nolint:gosec
	}

	client := pmg.NewAPIClient(cfg)

	// /version is a trivial REST sanity check (no model decoding —
	// avoids the int32 `mem` field overflow on big-RAM hosts).
	if _, resp, err := client.VersionAPI.VersionVersion(context.Background()).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "REST sanity check:", err)
		os.Exit(1)
	} else {
		fmt.Printf("Connected (insecure TLS): /version → %d\n", resp.StatusCode)
	}

	if os.Getenv("PMG_NODE") == "" || os.Getenv("PMG_VMID") == "" {
		fmt.Println("(skip terminal: set PMG_NODE and PMG_VMID to test the WebSocket leg)")
		return
	}
	vmid64, _ := strconv.ParseInt(os.Getenv("PMG_VMID"), 10, 32)
	target := pmg.Target{Kind: pmg.TargetKindQemu, Node: os.Getenv("PMG_NODE"), Vmid: int32(vmid64)}

	session, err := client.ConnectTerminal(context.Background(), target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "connect terminal:", err)
		os.Exit(1)
	}
	session.OnMessage = func(text string) { fmt.Print(text) }
	_ = session.Send("uname -a\n")
	time.Sleep(3 * time.Second)
	_ = session.Close()
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
