# Alertmanager

## Prerequisites

- A running instance of Ubuntu 22.04 LTS for setting up Alertmanager
- A running instance of Prometheus. See [here](https://github.com/aviate-labs/node-metrics) for more details.
- Root or sudo privileges
- Basic knowledge of the command line and Linux system administration
- A (dummy) Matrix account for a bot user

## Installation

Complete the following steps on newly created instance of Ubuntu 22.04 LTS

### **Step 1: Update the System**

Before installing any new software, it's advisable to update your system with the latest packages. This ensures you have the most recent security patches and software updates.

```bash
sudo apt update && sudo apt upgrade -y
```

### **Step 2: Create a Alertmanager User**

For security reasons, Alertmanager should not run as the root user. Create a user and group for the Alert Manager to allow permission only for the specific user.

```bash
sudo groupadd -f alertmanager
sudo useradd -g alertmanager --no-create-home --shell /bin/false alertmanager
```

### **Step 3: Create Alertmanager Directories**

Move the Alertmanager binaries and configuration files to the appropriate directories and set the correct ownership to ensure Alertmanager runs under the correct user.

```bash
sudo mkdir -p /etc/alertmanager/templates
sudo mkdir /var/lib/alertmanager
sudo chown alertmanager:alertmanager /etc/alertmanager
sudo chown alertmanager:alertmanager /var/lib/alertmanager
```

### Step 4: Download and Configure Alertmanager

Download the latest version of Alertmanager from the [official website](https://prometheus.io/download/). Replace the URL with the latest version if necessary. The example below uses version 0.27.0.

```bash
pushd /tmp
wget https://github.com/prometheus/alertmanager/releases/download/v0.27.0/alertmanager-0.27.0.linux-amd64.tar.gz
```

Make sure to verify the integrity of the downloaded file using the `sha256sum` command. You can find the checksums on the download page.

```bash
sha256sum alertmanager-0.27.0.linux-amd64.tar.gz
tar xvf alertmanager-0.27.0.linux-amd64.tar.gz
cd alertmanager-0.27.0.linux-amd64.tar.gz
```

Copy the `alertmanager` and `amtol` files in the `/usr/bin` directory and change the group and owner to `alertmanager`. As well as copy the configuration file `alertmanager.yml` to the `/etc` directory and change the owner and group name to `alertmanager`.

```bash
sudo cp alertmanager /usr/bin/
sudo cp amtool /usr/bin/
sudo chown alertmanager:alertmanager /usr/bin/alertmanager
sudo chown alertmanager:alertmanager /usr/bin/amtool
sudo cp alertmanager.yml /etc/alertmanager/alertmanager.yml
sudo chown alertmanager:alertmanager /etc/alertmanager/alertmanager.yml
popd
```

### Step 5: Create a Alertmanager Service

Create a systemd service file to manage the Alertmanager service. This file ensures Alertmanager starts automatically on boot and can be controlled using the `systemctl` command.

```bash
sudo nano /etc/systemd/system/alertmanager.service
```

Add the following content to the file:

```
[Unit]
Description=AlertManager
Wants=network-online.target
After=network-online.target

[Service]
User=alertmanager
Group=alertmanager
Type=simple
ExecStart=/usr/bin/alertmanager \
    --config.file /etc/alertmanager/alertmanager.yml \
    --storage.path /var/lib/alertmanager/

[Install]
WantedBy=multi-user.target
```

Reload the systemd daemon to apply the changes and start the Alertmanager service. Enable the service to start on boot.

```bash
sudo systemctl daemon-reload
sudo systemctl start alertmanager
sudo systemctl enable alertmanager
```

### **Step 6: Verify Alertmanager Installation**

Ensure that Alertmanager is running correctly by checking the status of the service.

```bash
sudo systemctl status alertmanager
```

You should also be able to access the Alertmanager web interface at `http://localhost:9093` locally or `http://<your-ip>:9093` from a remote machine.

# Prometheus

Prometheus rules are crucial for triggering alerts. These rules define the conditions under which Prometheus will send an alert to the Alertmanager. The Alertmanager then routes these alerts to the appropriate channels. More details will be provided later.

You can define multiple rules in YAML files according to your alerting needs. For demonstration purposes, we will create a rule that triggers an alert when a instance is not reachable.

## Instance Down Alert

On your existing Prometheus instance, create a `.yml` file for this rule

```bash
sudo nano /etc/prometheus/instance_down_rules.yml
```

Insert the following into the newly created file

```bash
- name: alert.rules
  rules:
  - alert: InstanceDown
    expr: up == 0
    for: 5s
    labels:
      severity: critical
      webhook_url: 'test-alertmanager'
    annotations:
      summary: "Instance down"
      description: "Instance with Node ID: {{ $labels.instance }} has been down for more than 5 seconds."
```

This is a basic example to demonstrate the creation of alert rules. You can further customize the alert rules based on your specific monitoring needs. For more information on customizing Prometheus alerting rules, please visit the [Prometheus alerting rules documentation](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)

## **Update Prometheus Configuration**

Once the rule has been created, it needs to be added to the Prometheus configuration, along with the details of the Alertmanager.

Open the Prometheus configuration file

```bash
sudo nano /etc/prometheus/prometheus.yml
```

Add `alerting` and `rule_files` as top level keys 

```yaml
alerting:
  alertmanagers:
    - static_configs:
        - targets:
            - <your-ip>:9093    # or localhost:9093 

rule_files:
  - "instance_down_rules.yml"
```

See the [prometheus-alerts.example.yml](./prometheus-alerts.example.yml) file for how the `prometheus.yml` should look.

## Restart Prometheus Service

After modifying the configurations, restart the Prometheus service using the following command.

```bash
sudo systemctl restart prometheus.service
```

Check the status of the service to ensure there are no errors in the above configurations using the following command.

```bash
sudo systemctl status prometheus.service
```

# Matrix Bot Setup

### **Step 1: Retrieve Matrix (bot) account information**

If you haven’t already, **[create a new user account](https://app.element.io/#/login)** using the Element client to be the bot on your Matrix home server. Note the bot user's access token, user ID, and home server URL using the instructions below.

- User ID: `Settings > General > Username (e.g. @botusername:matrix.org)`
- Homeserver url : `Settings > General > Help & About > Advanced > Homeserver`
- Access Token: A long lived access token will be needed so that your Matrix bot can send notifications in the background. 

  Copy and paste the following command in your terminal to get your access token. Make sure to replace `<home-server-url` , `<your-bot-user-id` and `your-bot-account-password`.
      
    ```bash
    curl -X POST "https://matrix-client.matrix.org/_matrix/client/v3/login" -H "Content-Type: application/json" -d '{"type":"m.login.password","identifier":{"type":"m.id.user","user":"<your-bot-user-id>"},"password":"<your-bot-account-password>"}'
    ```
    
  Your access token should be visible in the output with the this format: `syt_abced...` .
    

### **Step 2: Retrieve Matrix room information**

On your user account, create a room which you can invite the newly created bot using its user ID. Take note of the room ID: `Room options > Settings > Advanced > Internal room ID`

# Matrix Receiver

Receivers are the destinations for alerts after they have been processed by Alertmanager. To send alerts via the Matrix protocol, a Matrix receiver is required. We will use the [matrix-alertmanager-receiver by metio](https://github.com/metio/matrix-alertmanager-receiver). Alternatively, you can develop your own receiver.

### Step 1: Create a dedicated user and directory for the receiver

Create the `matrix-alertmanager-receiver` user that will run this service, as well as the `/etc/matrix-alertmanager-receiver` directory to store the necessary configuration files.

```bash
sudo useradd --no-create-home --shell /bin/false matrix-alertmanager-receiver
sudo mkdir /etc/matrix-alertmanager-receiver
```

### Step 2: Clone the repository

On the instance where you previously downloaded and configured Alertmanager, clone the `matrix-alertmanager-receiver` repository and navigate to its directory.

```bash
pushd /tmp
git clone https://github.com/metio/matrix-alertmanager-receiver.git
cd matrix-alertmanager-receiver
```

### Step 3: Configure the receiver

Copy the `config.sample.yaml` file to a `config.yaml` file

```bash
cp config.sample.yaml config.yaml
sudo nano config.yaml
```

Enter the details collected from element into the configuration file. Below is a minimal example:

```yaml
# configuration of the HTTP server
http:
  port: <port-on-your-instance> # e.g. 12345

# configuration for the Matrix connection
matrix:
  homeserver-url: "<your-homeserver-url>" # e.g. https://matrix-client.matrix.org
  user-id: "<your-bot-user-id>" # e.g. @some-username:matrix.org
  access-token: "<your-bot-access-token>" # e.g. syt_abc... 
  room-mapping:
	  simple-name: "<your-room-id>" # e.g. !qohfwef7qwerf:example.com
# configuration of templating features
templating:
  external-url-mapping:
    "alertmanager:9093": https://alertmanager.example.com
  computed-values:
    - values:
        color: white
    - values:
        color: orange
      when-matching-labels:
        severity: warning
    - values:
        color: red
      when-matching-labels:
        severity: critical
    - values:
        color: limegreen
      when-matching-status: resolved
  firing-template: '
    <p>
      <strong><font color="{{ .ComputedValues.color }}">{{ .Alert.Status | ToUpper }}</font></strong>
      {{ if .Alert.Labels.name }}
        {{ .Alert.Labels.name }}
      {{ else if .Alert.Labels.alertname }}
        {{ .Alert.Labels.alertname }}
      {{ end }}
      >>
      {{ if .Alert.Labels.severity }}
        {{ .Alert.Labels.severity | ToUpper }}:
      {{ end }}
      {{ if .Alert.Annotations.description }}
        {{ .Alert.Annotations.description }}
      {{ else if .Alert.Annotations.summary }}
        {{ .Alert.Annotations.summary }}
      {{ end }}
      >>
      {{ if .Alert.Annotations.runbook }}
        <a href="{{ .Alert.Annotations.runbook }}">Runbook</a> |
      {{ end }}
      {{ if .Alert.Annotations.dashboard }}
        <a href="{{ .Alert.Annotations.dashboard }}">Dashboard</a> |
      {{ end }}
      <a href="{{ .SilenceURL }}">Silence</a>
    </p>'
```

For full details on each key in the `config.yaml` file, please refer to the [repository’s documentation](https://github.com/metio/matrix-alertmanager-receiver?tab=readme-ov-file#configuration).

### Step 4: Build the Matrix Alertmanager receiver

In the `matrix-alertmanager-receiver` directory, execute the following command.

```bash
CGO_ENABLED=0 go build -o matrix-alertmanager-receiver
```

You may need to have, at least, Golang 1.21 installed.

```bash
sudo apt install golang-go
```

### Step 5: Create and configure the Matrix Alertmanager receiver service

Copy the `matrix-alertmanager-receiver` executable files to the `/usr/bin` directory and change the group and owner to `matrix-alertmanager-receiver`. As well as copy the configuration file `config.yaml` to the `/etc/matrix-alertmanager-receiver` directory and change the owner and group name to `matrix-alertmanager-receiver`.

```bash
sudo mv /tmp/matrix-alertmanager-receiver/matrix-alertmanager-receiver /usr/local/bin/
sudo mv /tmp/matrix-alertmanager-receiver/config.yml /etc/matrix-alertmanager-receiver/config.yml
sudo chown matrix-alertmanager-receiver:matrix-alertmanager-receiver /usr/local/bin/matrix-alertmanager-receiver
sudo chown matrix-alertmanager-receiver:matrix-alertmanager-receiver /etc/matrix-alertmanager-receiver/config.yaml
popd
```

Create a systemd service file to manage the `matrix-alertmanager-receiver` service. This file ensures `matrix-alertmanager-receiver` starts automatically on boot and can be controlled using the `systemctl` command.

```bash
sudo nano /etc/systemd/system/matrix-alertmanager-receiver.service
```

Add the following content to the file:

```bash
[Unit]
Description=Matrix Alertmanager Receiver Service
After=network.target

[Service]
User=matrix-alertmanager-receiver
Group=matrix-alertmanager-receiver
Type=simple
ExecStart=/usr/local/bin/matrix-alertmanager-receiver --config-path /etc/matrix-alertmanager-receiver/config.yaml

[Install]
WantedBy=multi-user.target
```

Reload the systemd daemon to apply the changes and start the `matrix-alertmanager-receiver` service. Enable the service to start on boot.

```bash
sudo systemctl daemon-reload
sudo systemctl enable matrix-alertmanager-receiver
sudo systemctl start matrix-alertmanager-receiver
```

### **Step 5: Verify Matrix Alertmanager receiver setup**

Ensure that `matrix-alermanager-receiver` is running correctly by checking the status of the service.

```bash
sudo systemctl status matrix-alertmanager-receiver
```

# Configure Alertmanager Routes and Receivers

Open the `alertmanager.yml` file

```bash
sudo nano /etc/alertmanager/alertmanager.yml
```

Add the routes and receivers to the file. Here is an example

```bash
global:
  resolve_timeout: 5m
  
route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'
  routes:
    - match:
        webhook_url: 'test-alertmanager'
      group_by: ['alertname']
      group_wait: 10s
      group_interval: 10s
      repeat_interval: 15s
      receiver: 'matrix'

receivers:
  - name: 'web.hook'
    webhook_configs:
      - url: 'http://127.0.0.1:5001/'
  - name: 'matrix'
    webhook_configs:
      - url: 'http://167.99.39.150:12345/alerts/simple-name'

inhibit_rules:
  - source_match:
      severity: 'critical'
    target_match:
      severity: 'warning'
    equal: ['alertname', 'dev', 'instance']
```

Restart the `alertmanager` service

```bash
sudo systemctl restart alertmanager
```