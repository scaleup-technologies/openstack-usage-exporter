# openstack-usage-exporter

## Motivation

As a cloud provider we need a reliable way of getting usage data. The [prometheus-openstack-exporter](https://github.com/canonical/prometheus-openstack-exporter) project can do this.

However, it is well known amongst OpenStack operators that the API is on the slow side. The openstack exporter puts a high workload on them and gathers a lot of data that is not always needed.

Usage data can also be retrieved via the database, but as a momentary value only. This is why we created this exporter.

```
MariaDB [nova]> select project_id, sum(vcpus) as total_vcpus from instances where deleted = 0 group by project_id;
+----------------------------------+-------------+
| project_id                       | total_vcpus |
+----------------------------------+-------------+
| 4679b3bfa8414681b2cde9412ef891ec |           2 |
| 70ac19f6b4504ccc8c2030a63d8da9bf |           7 |
| 7ad8fe83dd1e43ceb32126a88be1adcb |           4 |
+----------------------------------+-------------+
```

Using prometheus with a regular scrape interval, this can be used for metering purposes and more.

## Usage

```shell
go build

export BASE_DSN="dbuser:dbpass@tcp(localhost:3306)"
./openstack-usage-exporter

curl http://localhost:9143/metrics
```

Note: it is highly recommended to use a read-only user. Permissions must be granted to all affected databases (nova, cinder etc)

## Architecture

The exporter will run SQL queries on demand when queried. It is therefor important to consider the scrape interval to prevent high load on the database. For redundancy deploy the exporter on multiple hosts and add a load balancer (e.g. haproxy) in front, to only query one exporter at a time.

## Configuration

Configuration is done via enviroment variables.

Exporters can be enabled or disabled:

```shell
# Default values
NOVA_ENABLED=true
CINDER_ENABLED=true
DESIGNATE_ENABLED=true
MANILA_ENABLED=false
NEUTRON_ENABLED=true
OCTAVIA_ENABLED=true

# Routers returned by the Neutron Exporter are filtered by a specific external network ID.
# This is designed to only count the usage of routers which are connected to an external network.
NEUTRON_ROUTER_EXTERNAL_NETWORK_ID=5d8722dd-186c-4e32-a170-b216a04688dc
```
