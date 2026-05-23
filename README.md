# pmg-go

Go SDK for the Proxmox Mail Gateway API. Generated from
the upstream `apidoc.js` from Proxmox Mail Gateway via [openapi-generator-cli][gen] with custom
Mustache template overrides.

> **Not an official Proxmox project.** Community SDK derived from the
> upstream `apidoc.js`. Always verify against the upstream API viewer.
> <https://pmg.proxmox.com/>.

Requires Go ≥ 1.22.

## Install

```bash
go get github.com/client-api/pmg-go
```

## Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/client-api/pmg-go"
)

func main() {
    cfg := openapi.NewConfiguration()
    cfg.Servers[0].URL = "https://pmg1.example.com:8006/api2/json"
    cfg.AddDefaultHeader("Authorization", "PMGAPIToken=user@realm!tokenid=uuid-secret")

    pve := openapi.NewPmg(cfg)

    status, _, err := pve.QemuAPI.QemuVmStatus(context.Background(), "pmg1", 100).Execute()
    if err != nil {
        panic(err)
    }
    fmt.Printf("%+v\n", status)
}
```

`Pmg` is a brand alias for the generator's `*APIClient`, which already
exposes every per-tag API as a struct field
(`QemuAPI`, `LxcAPI`, `ClusterAPI`, …).

## Compound configs

PVE encodes many fields as CLI-style shorthand strings
(`net0=virtio,bridge=vmbr0,firewall=1`). Round-trip helpers are
emitted for every compound config schema:

```go
cfg := openapi.PveQemuNetConfig{
    Model:    "virtio",
    Bridge:   stringPtr("vmbr0"),
    Firewall: int32Ptr(1),
}
shorthand := cfg.ToShorthand()
// → "virtio,bridge=vmbr0,firewall=1"
```

## License

Apache 2.0 — see [LICENSE](./LICENSE).

[gen]: https://openapi-generator.tech
