# sparsnas_mqtt_consumer

sparsnas_mqtt_consumer is an application subscribing for decoded messages (by <https://github.com/tubalainen/sparsnas_decoder>) published to a MQTT topic and dispatching them forward to other data sources. Currently InfluxDB is the only supported output data source, but the code is design such that new data sources can be added easily.

Additionally the repo contains example Grafana dashboards for plotting the data collected data.

## How to use?

1. Build and run the consumer

```
go build && ./sparsnas_mqtt_consumer -influx-forward
```

Runtime configuration flags

```
Usage of ./sparsnas_mqtt_consumer:
  -version
     print version and exit

This application is configured via the environment. The following environment
variables can be used:

KEY                TYPE             DEFAULT                  REQUIRED    DESCRIPTION
MQTT_BROKER        String           localhost                            IP or hostaname to the MQTT broker
MQTT_PORT          Integer          1883                                 port to the MQTT broker
MQTT_USERNAME      String                                                username for MQTT broker
MQTT_PASSWORD      String                                                password for MQTT broker
MQTT_TOPIC         String           #
INFLUX_ADDR        String           http://localhost:8086                address to influxdb for storing measurements
INFLUX_DATABASE    String           sparsnas                             name of the influx database
INFLUX_FORWARD     True or False    false                                forward messages to influx
```

## Development

1. Fire up the backend services and setup the influx database

```
docker-compose up -d
./create-database.sh
```

2. Build and run the mock sparsnas_decoder publishing generated readings to MQTT

```
cd mock_sparsnas_decoder
go build
./mock_sparnas_decoder
```

3. Build and run the consumer

```
go build
./sparsnas_mqtt_consumer
```
