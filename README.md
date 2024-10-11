# f5ltm_exporter
F5 Local Traffic Management Device Exporter

F5 LTM Exporter is implemented via the multi-target exporter pattern.
By multi-target exporter pattern we refer to a specific design, in which:

- the exporter will get the target’s metrics via a network protocol.
- the exporter does not have to run on the machine the metrics are taken from.
- the exporter gets the targets and a query config string as parameters of Prometheus’ GET request.
- the exporter subsequently starts the scrape after getting Prometheus’ GET requests and once it is done with scraping.

When the exporter starts the scarape, it is performing following actions:
- POST /mgmt/shared/authn/login
- GET /mgmt/tm/ltm/pool/stats

## Getting Started
The project is developed in Go (1.23+).\
The repository is formatted for use in GoLand.

NOTE: The rest of this README assumes you are using GoLand.

## Prerequisites
Development requirements:
* GoLand.

## How to start
* Install GoLand .
* Open GoLand - Clone  project from Github

## Run System
* make start
* Open a web browser and navigate to `http://localhost:9143/probe?target=f5.somehost.com`

## Build System
* make build

## Push System to repository
* make deploy


## Environment
    HOST                (default binds to 0.0.0.0)
    PORT                (listening port, default 9143)
    F5_USER
    F5_PASS

## Installation
```console
sudo useradd --no-create-home --shell /bin/false f5ltmexporter
sudo mkdir /opt/f5ltm_exporter
cd /opt/f5ltm_exporter
sudo tar -xvf f5ltm_exporter_0.0.2_linux_amd64.tar.gz
sudo chmod 755 f5ltmexporterserver
sudo chown f5ltmexporter:f5ltmexporter /opt/f5ltm_exporter/*
sudo ln -s /opt/f5ltm_exporter/f5ltmexporterserver /usr/local/bin/f5ltmexporterserver

sudo tee /etc/systemd/system/f5ltm_exporter.service <<EOF
[Unit]
Description=F5 Exporter
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
WorkingDirectory=/opt/f5ltm_exporter
ExecStart=/usr/local/bin/f5ltmexporterserver

Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable --now f5ltm_exporter.service 
```

## Prometheus configuration
```yaml
    - job_name: 'f5ltm_exporter'
      metrics_path: /probe
      static_configs:
        - targets:
          - f5.somehost.com
      relabel_configs:
        - source_labels: [__address__]
          target_label: __param_target
        - source_labels: [__param_target]
          target_label: instance
        - target_label: __address__
          replacement: 127.0.0.1:9143
```

## Filebeat configuration
```console
- type: journald
  enabled: true
  pipeline: filebeat
  id: service-f5ltm-exporter
  include_matches.match:
    - _SYSTEMD_UNIT=f5ltm_exporter.service
  fields:
    type: f5ltm_exporter.server

  parsers:
    - ndjson:
      overwrite_keys: true
      add_error_key: true
      expand_keys: true
```