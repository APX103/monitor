# GPU server monitor

> A mini program for monitor GPU server

## stack

1. gopsutil
2. os/exec + nvidia-smi
3. influxdb

## monitor index

1. CPU Utilization
2. Memory Usage
3. GPU Utilization
4. GPU Memory Usage

## Usage

1. Install Influxdb
2. Setup Influxdb and get API token
3. Just build this program and run with parameters

## monitor's parameters

1. `bucket` : The bucket name of influxdb2
2. `org` : The org name of influxdb2
3. `token` : The API token of influxdb2
4. `url` : The url of influxdb2

It means. You can push your data to any influxdb2 the server can reach.
And, if any param goes wrong, you get no panic...at least now its true.
