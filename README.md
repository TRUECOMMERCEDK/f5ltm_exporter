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


## Enviroment
    HOST                (default binds to 0.0.0.0)
    PORT                (listening port, default 9143)
    F5_USER
    F5_PASS

## Prometheus configuration

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
