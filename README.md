# f5ltm_exporter
F5 Local Traffic Management Device Exporter

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
* Open a web browser and navigate to `http://localhost:9142/metrics`

## Build System
* make build

## Push System to repository
* make deploy


## Enviroment
    HOST                (default binds to 0.0.0.0)
    PORT                (listening port, default 9142)
    METRICS_PATH        (default /metrics)
    F5_USER
    F5_PASS
    F5_HOST

