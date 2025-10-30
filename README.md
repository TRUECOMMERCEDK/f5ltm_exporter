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
- Performs a **fresh login per scrape** and deletes the token immediately after use  
  (no token caching, preventing “maximum active tokens” errors)
- TLS support with configurable verification via `--tls-skip-verify`
- Exposes standard endpoints:
    - `/probe?target=<F5-IP>` – scrape endpoint used by Prometheus
    - `/metrics` – exporter self-metrics
    - `/healthz` – simple health check

---

## Command-Line Flags

| Flag                | Description                                   | Default     |
|---------------------|-----------------------------------------------|-------------|
| `--host`            | Address to bind the exporter                  | `127.0.0.1` |
| `--port`            | Port number to bind the exporter              | `9143`      |
| `--f5-user`         | F5 username (required)                        | –           |
| `--f5-pass`         | F5 password (required)                        | –           |
| `--tls-skip-verify` | Skip TLS verification (use only for testing)  | `false`     |
| `--log-format`      | Log format: `json` or `text`                  | `json`      |
| `--log-level`       | Log level: `debug`, `info`, `warn`, or `error`| `info`      |

Example:
```bash
./f5ltm_exporter   --f5-user=admin   --f5-pass=secret 
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
- iControl REST uses HTTPS; certificate verification can be disabled with --tls-skip-verify (for testing only).
- The exporter performs a **login and logout per scrape**, ensuring no long-lived sessions remain active.
- Credentials are passed via flags; consider using environment variables or a secrets manager for production.

---

## Architecture
```
Prometheus
   │
   ├── /probe?target=f5-a → Login → Scrape → Logout
   ├── /probe?target=f5-b → Login → Scrape → Logout
   └── /metrics → exporter self-metrics
```

Each scrape:
- Creates a new short-lived token
- Collects pool and sync metrics
- Logs out and deletes the token
No tokens are cached or reused between scrapes.
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
ExecStart=/opt/f5ltm_exporter/f5ltm_exporter --f5-user=admin --f5-pass=secret
Restart=always

[Install]
WantedBy=multi-user.target
```
## Performance Considerations
Since each scrape performs a **login and logout**, it will add slight latency (typically ~200–400 ms per scrape).
To avoid excessive load on the F5 device:
- Use a scrape interval of **30–60 seconds**.
- Avoid overly frequent multi-target scrapes in large clusters.
- You can run multiple exporters in parallel if needed for scaling.

---

## License
MIT License – see [LICENSE](LICENSE)
