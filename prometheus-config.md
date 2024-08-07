# Prometheus

The node exposes a Prometheus endpoint for scraping metrics. Follow the steps below to install and configure Prometheus
on your Ubuntu instance.

### Prerequisites

- A running instance of Ubuntu 22.04 LTS
- Root or sudo privileges
- Basic knowledge of the command line and Linux system administration

### Installation

#### Step 1: Update the System

Before installing any new software, it's advisable to update your system with the latest packages. This ensures you have
the most recent security patches and software updates.

```shell
sudo apt update && sudo apt upgrade -y
```

#### Step 2: Create a Prometheus User

For security reasons, Prometheus should not run as the root user. Create a dedicated user for Prometheus with restricted
permissions.

```shell
sudo useradd --no-create-home --shell /bin/false prometheus
```

#### Step 3: Download Prometheus

Download the latest version of Prometheus from the [official website](https://prometheus.io/download/). Replace the URL
with the latest version if
necessary. The example below uses version 2.45.5 LTS.

```shell
pushd /tmp
wget https://github.com/prometheus/prometheus/releases/download/v2.45.5/prometheus-2.45.5.linux-amd64.tar.gz
```

Make sure to verify the integrity of the downloaded file using the `sha256sum` command. You can find the checksums on
the download page.

```shell
sha256sum prometheus-2.45.5.linux-amd64.tar.gz 
tar xvf prometheus-2.45.5.linux-amd64.tar.gz
popd
```

#### Step 4: Configure Prometheus

Move the Prometheus binaries and configuration files to the appropriate directories and set the correct ownership to
ensure Prometheus runs under the correct user.

```shell
sudo mv /tmp/prometheus-2.45.5.linux-amd64/prometheus /usr/local/bin/
sudo mv /tmp/prometheus-2.45.5.linux-amd64/prometheus.yml /etc/prometheus/prometheus.yml

sudo chown prometheus:prometheus /usr/local/bin/prometheus

sudo mkdir /var/lib/prometheus
sudo chown prometheus:prometheus /var/lib/prometheus
```

You can now edit the Prometheus configuration file to scrape metrics from the node. You can find an
example [here](./prometheus.example.yml).

```shell
sudo nano /etc/prometheus/prometheus.yml
```

#### Step 5: Create a Prometheus Service

Create a systemd service file to manage the Prometheus service. This file ensures Prometheus starts automatically on
boot and can be controlled using the `systemctl` command.

```shell
sudo nano /etc/systemd/system/prometheus.service
```

Add the following content to the file:

```ini
[Unit]
Description=Prometheus Monitoring
Wants=network-online.target
After=network-online.target

[Service]
User=prometheus
Group=prometheus
Type=simple
ExecStart=/usr/local/bin/prometheus \
    --config.file=/etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/var/lib/prometheus/

[Install]
WantedBy=multi-user.target
```

Reload the systemd daemon to apply the changes and start the Prometheus service. Enable the service to start on boot.

```shell
sudo systemctl daemon-reload
sudo systemctl start prometheus
sudo systemctl enable prometheus
```

#### Step 6: Verify Prometheus Installation

Ensure that Prometheus is running correctly by checking the status of the service.

```shell
sudo systemctl status prometheus
```

You should also be able to access the Prometheus web interface at `http://localhost:9090` locally
or `http://<your-ip>:9090` from a remote machine.

## IPv6 Connectivity

To access node metrics over IPv6, you need to enable IPv6 connectivity on your machine. Depending on your cloud
provider, additional steps might be required to enable IPv6 on your instance (
e.g. [Enable IPv6 on Digital Ocean Droplets]([url](https://docs.digitalocean.com/products/networking/ipv6/how-to/enable/#on-existing-droplets))).

This section provides a general guide on how to enable IPv6 on an Ubuntu instance. Refer to your cloud provider's
documentation for specific instructions.

### Enabling IPv6 on Ubuntu

Edit the `/etc/netplan/50-cloud-init.yaml` file to enable IPv6 connectivity. This file configures networking on your
Ubuntu instance.

```shell
sudo nano /etc/netplan/50-cloud-init.yaml
```

Add or modify the configuration to include IPv6 settings, then apply the changes using:

```shell
sudo netplan apply
```

Following these steps, you will have a fully functioning Prometheus setup to monitor node metrics efficiently. You can now procede to [alertmanager-matrix-config.md](./alertmanager-matrix-config.md) to setup Alertmanager and receive alerts using the public node metrics.