# F5 LTM Exporter

## Overview
`f5ltm_exporter` exposes performance and HA metrics from **F5 BIG-IP LTM** devices for **Prometheus**.  
It uses the F5 iControl REST API to collect pool and sync-status information.

This exporter is **multi-target**:  
Prometheus passes the target F5 device via a `target` query parameter on each scrape.

---

## Features
- Collects **LTM pool statistics** and **sync-status**
- Supports multiple F5 devices dynamically (`?target=` parameter)
- Reuses API tokens automatically (no re-login storms)
- TLS support with `InsecureSkipVerify` for lab/test use
- Exposes standard endpoints:
    - `/probe?target=<F5-IP>` – scrape endpoint used by Prometheus
    - `/metrics` – exporter self-metrics
    - `/healthz` – simple health check

---

## Command-Line Flags

| Flag | Description | Default |
|------|--------------|----------|
| `--host` | Address to bind the exporter | `0.0.0.0` |
| `--port` | Port number to bind the exporter | `9143` |
| `--f5-user` | F5 username (required) | – |
| `--f5-pass` | F5 password (required) | – |

Example:
```bash
./f5ltm_exporter   --f5-user admin   --f5-pass secret   --host 0.0.0.0   --port 9143
```

Then open:
```
http://localhost:9143/probe?target=10.0.0.1
```

---

## Prometheus Configuration

```yaml
scrape_configs:
  - job_name: "f5ltm"
    metrics_path: /probe
    static_configs:
      - targets:
          - 10.0.0.1
          - 10.0.0.2
    params:
      module: [default]
```

Prometheus will call:
```
http://<exporter-host>:9143/probe?target=10.0.0.1
```

---

## Security Notes
- iControl REST uses HTTPS; this exporter skips certificate verification by default (`InsecureSkipVerify=true`).  
  You can harden it later with proper CA handling.
- Credentials are passed via flags; consider using environment variables or a secrets manager for production.

---

## Architecture
```
Prometheus
   │
   ├── /probe?target=f5-a → token cached in memory for "f5-a"
   ├── /probe?target=f5-b → token cached separately
   └── /metrics → exporter self-metrics
```

Each F5 host has:
- Independent login and cached token
- Automatic token reuse until expiry (~10h)
- On-demand refresh when expired

---

## Example Output

```
# HELP f5_pool_active_members Number of active members in each pool
# TYPE f5_pool_active_members gauge
f5_pool_active_members{pool="/Common/web_pool"} 4
f5_pool_active_members{pool="/Common/api_pool"} 2

# HELP f5_sync_status Indicates whether the device group is In Sync (1) or Out of Sync (0)
# TYPE f5_sync_status gauge
f5_sync_status 1
```

---

## Systemd Example

```ini
[Unit]
Description=F5 LTM Exporter
After=network.target

[Service]
User=f5ltm
ExecStart=/opt/f5ltm_exporter/f5ltm_exporter   --f5-user=admin   --f5-pass=secret
Restart=always

[Install]
WantedBy=multi-user.target
```

---

## License
MIT License – see [LICENSE](LICENSE)
