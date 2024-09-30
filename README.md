# Node Metrics

This document provides a detailed guide on how to set up and use node metrics for monitoring a node's performance. The
metrics are collected by the node and can be accessed through its REST API. Prometheus will be used to create a time series database from the data returned by the public node metrics, and Alertmanager will be set up and configured to send alert notifications via Matrix.

For a more comprehensive observability
stack, refer to the official [observability stack](https://github.com/dfinity/ic-observability-stack) repository.

## Getting Started

### Generating a Prometheus Configuration

To generate a Prometheus configuration file, you can use the golang script provided in this repository. The script
generates a configuration file based on a [node provider identifier](https://dashboard.internetcomputer.org/providers).

Generate a configuration file by running the following command:

```shell
go run main.go --node-provider-id="rbn2y-6vfsb-gv35j-4cyvy-pzbdu-e5aum-jzjg6-5b4n5-vuguf-ycubq-zae" --use-dashboard
```

The `--use-dashboard` flag is optional and generates a configuration file based on the node provider's dashboard. You
can also run the script without the flag to generate a configuration file based on the registry what will be fetched
from the main-net registry directly.

The `--out` flag can be used to specify the output file path. By default, the configuration file is saved in the current
directory as `prometheus.yml`.

This Prometheus configuration file can be used in guides mentioned in the next section.

### Setting up Prometheus and Alertmanager
1. Follow the instructions in [prometheus-config.md](./prometheus-config.md) to set up Prometheus
2. Follow the insturctions in [alertmanager-matrix-config.md](./alertmanager-matrix-config.md) to set up Alertmanager with Matrix.

### References

- [Forum: Node Metrics](https://forum.dfinity.org/t/public-internet-computer-ic-node-metrics-available-now/32961)
